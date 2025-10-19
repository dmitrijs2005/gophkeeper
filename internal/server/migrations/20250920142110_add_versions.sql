-- +goose Up
-- +goose StatementBegin
ALTER TABLE entries ADD COLUMN version BIGINT NOT NULL DEFAULT 0;
ALTER TABLE entries ADD COLUMN deleted BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN current_version BIGINT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN current_version
ALTER TABLE entries DROP COLUMN deleted 
ALTER TABLE entries DROP COLUMN version 
-- +goose StatementEnd
