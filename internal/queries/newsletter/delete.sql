-- internal/queries/newsletter/delete.sql
DELETE FROM newsletters
WHERE id = $1 AND editor_id = $2; 