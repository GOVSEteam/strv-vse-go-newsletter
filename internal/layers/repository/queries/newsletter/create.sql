-- internal/queries/newsletter/create.sql
INSERT INTO newsletters (editor_id, name, description)
VALUES ($1, $2, $3)
RETURNING id, editor_id, name, description, created_at, updated_at; 