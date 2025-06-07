-- internal/queries/editor/get_by_id.sql
SELECT id, firebase_uid, email, created_at, updated_at
FROM editors
WHERE id = $1; 