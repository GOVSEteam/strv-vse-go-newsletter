-- internal/queries/post/mark_as_published.sql
UPDATE posts
SET published_at = $1, updated_at = $2
WHERE id = $3
RETURNING id, newsletter_id, title, content, published_at, created_at, updated_at; 