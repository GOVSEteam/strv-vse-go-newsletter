-- internal/queries/editor/get_by_firebase_uid.sql
SELECT id, firebase_uid, email, created_at, updated_at
FROM editors
WHERE firebase_uid = $1; 