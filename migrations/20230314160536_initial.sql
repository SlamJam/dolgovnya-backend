-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    title TEXT NOT NULL CHECK (LENGTH(title) > 0) UNIQUE
);

CREATE TABLE base_accounting (
    -- BIGSERIAL удобен при наследовании таблиц
    id BIGSERIAL PRIMARY KEY
);

CREATE TABLE accounting_entries (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    owner_accounting BIGINT NOT NULL REFERENCES base_accounting(id) ON DELETE CASCADE,
    user_from BIGINT NOT NULL REFERENCES users(id),
    user_to BIGINT NOT NULL REFERENCES users(id),
    value BIGINT CHECK (value <> 0),
    CONSTRAINT no_self_to_self CHECK (user_from != user_to)
);

-- id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
