BEGIN;

-- 1. Node Template for the Consignment (Macro) Workflow.
-- This node acts as a pointer to the Micro Workflow.
INSERT INTO workflow_node_templates (id, name, description, type, config, depends_on)
VALUES (
    'template-simple-form-v1',
    'Simple Form Task (Macro Node)',
    'A task node in the consignment workflow that delegates to a micro-workflow',
    'SIMPLE_FORM',
    '{
        "renderSchemaId": "simple-form-micro-v2",
        "microWorkflowId": "simple-form-micro-v2"
    }',
    '[]'
) ON CONFLICT (id) DO NOTHING;

-- 2. The Macro Workflow Definition (e.g., the Consignment process).
INSERT INTO workflow_template_v2 (id, name, version, workflow_definition)
VALUES
(
    'simple-form-macro-v1',
    'Simple Form Macro Workflow (Test)',
    '1',
    '{
        "id": "simple-form-macro-v1",
        "name": "Simple Form Macro Workflow (Test)",
        "version": 1,
        "nodes": [
            { "id": "node_0_start", "type": "START" },
            { "id": "node_1_form", "type": "TASK", "task_template_id": "template-simple-form-v1" },
            { "id": "node_2_end", "type": "END" }
        ],
        "edges": [
            { "id": "e_start", "source_id": "node_0_start", "target_id": "node_1_form" },
            { "id": "e_form_to_end", "source_id": "node_1_form", "target_id": "node_2_end" }
        ]
    }'::jsonb
) ON CONFLICT (id) DO NOTHING;

-- 3. Map the Macro workflow to an HS code for testing.
INSERT INTO workflow_template_maps_v2 (id, hs_code_id, consignment_flow, workflow_template_id)
VALUES
    (
        '55555555-5555-4000-8000-000000000001',
        '90b06747-cfa7-486b-a084-eaa1fc95595e',
        'EXPORT',
        'simple-form-macro-v1'
    ) ON CONFLICT (id) DO NOTHING;

COMMIT;
