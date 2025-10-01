-- +goose Up
-- +goose StatementBegin
ALTER TABLE entries ADD COLUMN pending NOT NULL DEFAULT 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE entries DROP COLUMN pending;
-- +goose StatementEnd
