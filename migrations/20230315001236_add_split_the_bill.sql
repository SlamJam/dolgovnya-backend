-- История разбития счёта --

-- +goose Up
-- +goose StatementBegin
CREATE TABLE accounting_split_the_bill (
    owner_id BIGINT NOT NULL REFERENCES users(id),
    bill jsonb
) INHERITS (base_accounting);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
