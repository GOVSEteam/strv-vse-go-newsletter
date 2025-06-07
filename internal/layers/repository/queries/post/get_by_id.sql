-- internal/queries/post/get_by_id.sql
SELECT id, newsletter_id, title, content, published_at, created_at, updated_at
FROM posts
WHERE id = $1; 