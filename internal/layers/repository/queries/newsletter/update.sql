-- internal/queries/newsletter/update.sql
UPDATE newsletters
SET name = COALESCE($1, name), description = COALESCE($2, description), updated_at = NOW()
WHERE id = $3 AND editor_id = $4
RETURNING id, editor_id, name, description, created_at, updated_at; 