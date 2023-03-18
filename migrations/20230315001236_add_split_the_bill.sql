-- История разбития счёта --

-- +goose Up
-- +goose StatementBegin
CREATE TABLE accounting_split_the_bill (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL,
    owning_object_id BIGINT NOT NULL,
    FOREIGN KEY (owning_object_id, user_id) REFERENCES owner_objects(id, user_id) ON DELETE CASCADE ON UPDATE CASCADE,

    schema_version INTEGER NOT NULL,
    bill jsonb
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
