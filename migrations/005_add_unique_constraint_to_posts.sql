-- +goose Up
-- This migration adds a composite unique constraint to the posts table.
-- This ensures that a post title is unique within a single newsletter.
ALTER TABLE posts
ADD CONSTRAINT posts_newsletter_id_title_key UNIQUE (newsletter_id, title);

-- +goose Down
-- This migration removes the composite unique constraint from the posts table.
ALTER TABLE posts
DROP CONSTRAINT posts_newsletter_id_title_key;
