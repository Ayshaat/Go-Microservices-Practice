-- +goose Up
-- +goose StatementBegin
ALTER TABLE cart_items
    ADD COLUMN price NUMERIC(10, 2) NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE cart_items
    DROP COLUMN price;
-- +goose StatementEnd
