package executors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/OpenNSW/nsw/internal/payments"
	"github.com/shopspring/decimal"
)

// NewPaymentInitiationExecutor returns an executor for the PAYMENT_INITIATION action.
func NewPaymentInitiationExecutor(ps payments.PaymentService) func(context.Context, string, map[string]any, json.RawMessage) (map[string]any, error) {
	return func(ctx context.Context, taskId string, inputs map[string]any, config json.RawMessage) (map[string]any, error) {
		slog.InfoContext(ctx, "executing payment initiation", "taskId", taskId)

		// 1. Prepare checkout session request.
		// In a production setup, we would read the amount and currency from the task's context.
		// For now, we use a placeholder.
		req := payments.CreateCheckoutRequest{
			ReferenceNumber: "REF-" + taskId[:8],
			Amount:          decimal.NewFromFloat(100.0), // Placeholder
			Currency:        "LKR",
			ExpiresAt:       time.Now().Add(time.Hour),
			Metadata:        map[string]string{"task_id": taskId},
		}

		// 2. Call the payment gateway service.
		resp, err := ps.CreateCheckoutSession(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to create checkout session: %w", err)
		}

		// 3. Return the data to be merged into the task context.
		return map[string]any{
			"checkoutUrl":      resp.CheckoutURL,
			"gatewaySessionId": resp.SessionID,
			"status":           "PENDING",
			"initiatedAt":      time.Now().Format(time.RFC3339),
		}, nil
	}
}
