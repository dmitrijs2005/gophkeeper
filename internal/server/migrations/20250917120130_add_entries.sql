-- +goose Up
-- +goose StatementBegin
CREATE TABLE entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,              -- например "Google account"
    type TEXT NOT NULL,               -- "login", "note", "card" ...
    encrypted_data BYTEA NOT NULL,    -- ciphertext
    nonce BYTEA NOT NULL,             -- nonce для AES-GCM
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE entries;
-- +goose StatementEnd
