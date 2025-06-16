-- +goose Up
-- This migration adds a composite unique constraint to the newsletters table.
-- This ensures that a single editor cannot have two newsletters with the same name.
ALTER TABLE newsletters
ADD CONSTRAINT newsletters_editor_id_name_key UNIQUE (editor_id, name);

-- +goose Down
-- This migration removes the composite unique constraint from the newsletters table.
ALTER TABLE newsletters
DROP CONSTRAINT newsletters_editor_id_name_key;
