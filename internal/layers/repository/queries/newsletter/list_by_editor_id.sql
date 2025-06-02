-- internal/queries/newsletter/list_by_editor_id.sql
SELECT id, editor_id, name, description, created_at, updated_at
FROM newsletters
WHERE editor_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3; 