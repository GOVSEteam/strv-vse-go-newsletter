-- internal/queries/newsletter/get_by_id.sql
SELECT id, editor_id, name, description, created_at, updated_at
FROM newsletters
WHERE id = $1; 