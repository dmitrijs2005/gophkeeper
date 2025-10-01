-- +goose Up
-- +goose StatementBegin
CREATE TABLE entries (
    id UUID PRIMARY KEY NOT NULL,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    overview BYTEA NOT NULL,
    nonce_overview BYTEA NOT NULL,
    details BYTEA NOT NULL,
    nonce_details BYTEA NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE entries;
-- +goose StatementEnd
