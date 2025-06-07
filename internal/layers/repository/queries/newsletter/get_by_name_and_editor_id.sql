-- internal/queries/newsletter/get_by_name_and_editor_id.sql
SELECT id, editor_id, name, description, created_at, updated_at
FROM newsletters
WHERE name = $1 AND editor_id = $2; 