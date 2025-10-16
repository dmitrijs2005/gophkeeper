-- +goose Up
-- +goose StatementBegin
CREATE TABLE files (
    entry_id            UUID NOT NULL REFERENCES entries(id) UNIQUE,
    user_id             UUID NOT NULL REFERENCES users(id),
    version             BIGINT NOT NULL DEFAULT 0,
    storage_key         TEXT NOT NULL UNIQUE,
    encrypted_file_key  BYTEA NOT NULL,
    nonce               BYTEA NOT NULL,
    upload_status       TEXT NOT NULL DEFAULT 'pending',
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE files;
-- +goose StatementEnd
