-- internal/queries/newsletter/get_by_id_and_editor_id.sql
SELECT id, editor_id, name, description, created_at, updated_at
FROM newsletters
WHERE id = $1 AND editor_id = $2; 