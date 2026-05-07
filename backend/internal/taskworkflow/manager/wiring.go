package manager

import (
	"context"
	"fmt"
	"log/slog"

	workflowmanager "github.com/OpenNSW/go-temporal-workflow"
	"github.com/OpenNSW/nsw/internal/payments"
	"github.com/OpenNSW/nsw/internal/taskworkflow/manager/executors"
	"github.com/OpenNSW/nsw/internal/taskworkflow/persistence"
	"github.com/OpenNSW/nsw/internal/workflow/service"
	"github.com/OpenNSW/nsw/pkg/remote"
	"go.temporal.io/sdk/client"
	"gorm.io/gorm"
)

const microWorkflowTaskQueue = "TASK_MICRO_WORKFLOW_QUEUE"

// WireTaskManagerWithCleanup initializes the TaskWorkflowManager with its Temporal workers.
func WireTaskManagerWithCleanup(
	tc client.Client,
	db *gorm.DB,
	templateService service.TemplateProvider,
	paymentService payments.PaymentService,
	remoteManager *remote.Manager,
	macroCompletionHandler func(ctx context.Context, task *persistence.TaskWorkflowTask, outputs map[string]any) error,
) (TaskWorkflowManager, func() error) {
	store, _ := persistence.NewTaskWorkflowStore(db)
	commandStore, _ := persistence.NewTaskWorkflowCommandStore(db)

	var manager TaskWorkflowManager

	activationHandler := func(payload workflowmanager.TaskPayload) error {
		// This handler executes atomic nodes inside the Micro-Workflow.
		return handleAtomicTaskActivation(context.Background(), manager.(*taskWorkflowManager), payload)
	}

	completionHandler := func(taskID string, finalContext map[string]any) error {
		// When a Micro-Workflow completes, notify the Macro-Workflow.
		task, err := store.GetByTaskID(taskID)
		if err != nil {
			return fmt.Errorf("failed to get task for completion: %w", err)
		}
		return macroCompletionHandler(context.Background(), task, finalContext)
	}

	manager = NewTaskWorkflowManager(tc, microWorkflowTaskQueue, store, commandStore, templateService, activationHandler, completionHandler)

	// 1. Register Action Executors (Externalized Business Logic)
	m := manager.(*taskWorkflowManager)
	m.RegisterExecutor(string(ActionDataSubmission), executors.FormSubmissionExecutor)
	m.RegisterExecutor("OGA_API_CALL", executors.NewRemoteApiExecutor(remoteManager))
	m.RegisterExecutor(string(ActionPaymentInitiation), executors.NewPaymentInitiationExecutor(paymentService))

	// Start the worker
	if err := m.temporalManager.StartWorker(); err != nil {
		slog.Error("failed to start micro-workflow worker", "error", err)
	}

	cleanup := func() error {
		manager.(*taskWorkflowManager).temporalManager.StopWorker()
		return nil
	}

	return manager, cleanup
}
