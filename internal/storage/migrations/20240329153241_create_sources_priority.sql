-- +goose Up
-- +goose StatementBegin
ALTER TABLE sources ADD COLUMN priority INT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sources DROP COLUMN priority;
-- +goose StatementEnd