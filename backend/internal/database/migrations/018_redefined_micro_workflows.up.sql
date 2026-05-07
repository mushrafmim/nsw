BEGIN;

-- ============================================================================
-- 1. Atomic Node Templates for Micro-Workflows
-- These are reusable building blocks for ANY task graph.
-- ============================================================================

-- USER_INPUT building block
INSERT INTO workflow_node_templates (id, name, description, type, config, depends_on)
VALUES (
    'atomic-user-submission-v1',
    'User Data Entry',
    'Waiting for the trader or CHA to fill out and submit a form',
    'ATOMIC_STEP',
    '{
        "executionType": "WAIT",
        "action": "DATA_SUBMISSION",
        "pluginState": "AWAITING_SUBMISSION",
        "commands": [
            { "command": "SAVE_AS_DRAFT", "label": "Save Draft", "writeTo": "trader:form", "pluginState": "DRAFT"      },
            { "command": "SUBMIT_FORM",   "label": "Submit",      "writeTo": "trader:form", "pluginState": "SUBMITTED"  }
        ]
    }',
    '[]'
) ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config;

-- OGA_API building block
INSERT INTO workflow_node_templates (id, name, description, type, config, depends_on)
VALUES (
    'atomic-oga-api-call-v1',
    'Agency API Integration',
    'Automated background call to an Other Government Agency (OGA)',
    'ATOMIC_STEP',
    '{
        "executionType": "AUTO",
        "action": "OGA_API_CALL",
        "pluginState": "AWAITING_OGA_REVIEW",
        "serviceId": "npqs",
        "taskCode": "MOA_NPQS_PHYTOSANITARY_001",
        "endpoint": "/api/oga/inject",
        "template": {
            "taskId":        "${__runtime.taskId}",
            "consignmentId": "${__runtime.consignmentId}",
            "trader":        "${trader:form.businessName}",
            "items":         "${trader:form.items}"
        }
    }',
    '[]'
) ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config;

-- OGA_REVIEW building block
INSERT INTO workflow_node_templates (id, name, description, type, config, depends_on)
VALUES (
    'atomic-oga-approval-v1',
    'Agency Review Wait',
    'Waiting for the agency officer to approve or reject the submission',
    'ATOMIC_STEP',
    '{
        "executionType": "WAIT",
        "action": "OGA_VERIFICATION",
        "pluginState": "OGA_ACKNOWLEDGED",
        "commands": [
            { "command": "OGA_VERIFICATION", "label": "Approve",       "writeTo": "oga_review", "pluginState": "OGA_APPROVED"           },
            { "command": "REJECT",           "label": "Reject",         "writeTo": "oga_review", "pluginState": "OGA_REJECTED"           },
            { "command": "REQUEST_REWORK",   "label": "Request Rework", "writeTo": "oga_review", "pluginState": "OGA_FEEDBACK_REQUESTED" }
        ]
    }',
    '[]'
) ON CONFLICT (id) DO UPDATE SET config = EXCLUDED.config;


-- ============================================================================
-- 2. Redefined Simple Form Micro-Workflow
-- ============================================================================

INSERT INTO workflow_template_v2 (id, name, version, workflow_definition)
VALUES
(
    'simple-form-micro-v2',
    'Simple Form Micro Workflow (New Architecture)',
    '2',
    '{
        "id": "simple-form-micro-v2",
        "name": "Simple Form Micro Workflow",
        "version": 2,
        "nodes": [
            { "id": "START_NODE", "type": "START" },
            {
                "id": "USER_INPUT",
                "type": "TASK",
                "task_template_id": "atomic-user-submission-v1",
                "input_mapping": {
                    "__runtime": "__runtime"
                },
                "output_mapping": {
                    "trader:form": "trader:form"
                }
            },
            {
                "id": "API_CALL",
                "type": "TASK",
                "task_template_id": "atomic-oga-api-call-v1",
                "input_mapping": {
                    "__runtime":   "__runtime",
                    "trader:form": "trader:form"
                },
                "output_mapping": {}
            },
            {
                "id": "OGA_REVIEW",
                "type": "TASK",
                "task_template_id": "atomic-oga-approval-v1",
                "input_mapping": {
                    "__runtime": "__runtime"
                },
                "output_mapping": {
                    "oga_review": "oga_review"
                }
            },
            { "id": "END_NODE", "type": "END" }
        ],
        "edges": [
            { "id": "e1", "source_id": "START_NODE", "target_id": "USER_INPUT" },
            { "id": "e2", "source_id": "USER_INPUT", "target_id": "API_CALL"   },
            { "id": "e3", "source_id": "API_CALL",   "target_id": "OGA_REVIEW" },
            { "id": "e4", "source_id": "OGA_REVIEW", "target_id": "END_NODE"   }
        ]
    }'::jsonb
) ON CONFLICT (id) DO UPDATE SET workflow_definition = EXCLUDED.workflow_definition;

COMMIT;