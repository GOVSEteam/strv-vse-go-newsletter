-- internal/queries/newsletter/count_by_editor_id.sql
SELECT COUNT(*)
FROM newsletters
WHERE editor_id = $1; 