-- +goose Up
-- +goose StatementBegin
ALTER TABLE entries ADD COLUMN is_file NOT NULL DEFAULT 0;

-- Create files table for binary files
CREATE TABLE files (
    entry_id            TEXT PRIMARY KEY,
    encrypted_file_key  BLOB NOT NULL,
    nonce               BLOB NOT NULL,
    local_path          TEXT,
    upload_status       TEXT NOT NULL DEFAULT 'pending',
    deleted             INTEGER NOT NULL DEFAULT 0
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE files;
ALTER TABLE entries DROP COLUMN is_file;
-- +goose StatementEnd
