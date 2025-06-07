-- internal/queries/post/create.sql
INSERT INTO posts (id, newsletter_id, title, content, published_at, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, newsletter_id, title, content, published_at, created_at, updated_at; 