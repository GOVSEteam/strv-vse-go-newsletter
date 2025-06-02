-- internal/queries/post/update.sql
UPDATE posts
SET title = $1, content = $2, published_at = $3, updated_at = $4
WHERE id = $5
RETURNING id, newsletter_id, title, content, published_at, created_at, updated_at; 