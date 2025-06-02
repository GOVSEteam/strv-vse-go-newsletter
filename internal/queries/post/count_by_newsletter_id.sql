-- internal/queries/post/count_by_newsletter_id.sql
SELECT COUNT(*)
FROM posts
WHERE newsletter_id = $1; 