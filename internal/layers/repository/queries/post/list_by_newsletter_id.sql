-- internal/queries/post/list_by_newsletter_id.sql
SELECT id, newsletter_id, title, content, published_at, created_at, updated_at
FROM posts
WHERE newsletter_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3; 