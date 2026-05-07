BEGIN;

DELETE FROM workflow_template_maps_v2 WHERE id = '55555555-5555-4000-8000-000000000001';
DELETE FROM workflow_template_v2 WHERE id = 'simple-form-macro-v1';
DELETE FROM workflow_node_templates WHERE id = 'template-simple-form-v1';

COMMIT;
