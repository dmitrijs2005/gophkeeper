-- +goose Up
-- +goose StatementBegin
-- Create metadata table for salt/verifier/user_id/last_version
CREATE TABLE IF NOT EXISTS metadata (
    key TEXT PRIMARY KEY,
    value BLOB NOT NULL
);

-- Create entries table for overview/details
CREATE TABLE IF NOT EXISTS entries (
    id TEXT PRIMARY KEY,
    version BIGINT NOT NULL DEFAULT 0,
    deleted INTEGER NOT NULL DEFAULT 0, 

    overview BLOB NOT NULL,
    nonce_overview BLOB NOT NULL,

    details BLOB NOT NULL,
    nonce_details BLOB NOT NULL,

    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS entries;
DROP TABLE IF EXISTS metadata;
-- +goose StatementEnd