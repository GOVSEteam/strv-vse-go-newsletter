-- +goose Up
CREATE TABLE IF NOT EXISTS editors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create trigger function to automatically update updated_at field
CREATE OR REPLACE FUNCTION update_editors_updated_at()
RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = now(); RETURN NEW; END; $$ LANGUAGE plpgsql;

-- Create trigger to call the function before each update
CREATE TRIGGER trigger_editors_updated_at
    BEFORE UPDATE ON editors
    FOR EACH ROW
    EXECUTE FUNCTION update_editors_updated_at();

-- +goose Down
DROP TRIGGER IF EXISTS trigger_editors_updated_at ON editors;
DROP FUNCTION IF EXISTS update_editors_updated_at();
DROP TABLE IF EXISTS editors;
