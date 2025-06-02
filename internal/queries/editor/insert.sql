-- internal/queries/editor/insert.sql
INSERT INTO editors (firebase_uid, email)
VALUES ($1, $2)
RETURNING id, firebase_uid, email, created_at, updated_at; 