BEGIN;

DELETE FROM workflow_template_v2 WHERE id = 'simple-form-micro-v2';
DELETE FROM workflow_node_templates WHERE id IN (
    'atomic-user-submission-v1',
    'atomic-oga-api-call-v1',
    'atomic-oga-approval-v1'
);

COMMIT;
