-- internal/queries/post/delete.sql
DELETE FROM posts
WHERE id = $1; 