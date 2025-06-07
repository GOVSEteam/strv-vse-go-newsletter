-- +goose Up
CREATE TABLE IF NOT EXISTS newsletters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    editor_id UUID NOT NULL REFERENCES editors(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create trigger function to automatically update updated_at field
CREATE OR REPLACE FUNCTION update_newsletters_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to call the function before each update
CREATE TRIGGER trigger_newsletters_updated_at
    BEFORE UPDATE ON newsletters
    FOR EACH ROW
    EXECUTE FUNCTION update_newsletters_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trigger_newsletters_updated_at ON newsletters;
DROP FUNCTION IF EXISTS update_newsletters_updated_at();
DROP TABLE IF EXISTS newsletters; 