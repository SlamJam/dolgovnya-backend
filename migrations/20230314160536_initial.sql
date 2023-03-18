-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    title TEXT NOT NULL CHECK (LENGTH(title) > 0) UNIQUE
);

CREATE TABLE owner_objects (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    UNIQUE (id, user_id)
);

CREATE TABLE accounting_entries (
    id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL,
    owning_object_id BIGINT NOT NULL,
    FOREIGN KEY (owning_object_id, user_id) REFERENCES owner_objects(id, user_id) ON DELETE CASCADE ON UPDATE CASCADE,

    user_from BIGINT NOT NULL REFERENCES users(id),
    user_to BIGINT NOT NULL REFERENCES users(id),
    amount DECIMAL(14,2) CHECK (amount <> 0),
    CONSTRAINT no_self_to_self CHECK (user_from != user_to)
);

-- id BIGINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
