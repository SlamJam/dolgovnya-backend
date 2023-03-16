-- История разбития счёта --

-- +goose Up
-- +goose StatementBegin

-- какую p/q часть потребил пользователь user_id
CREATE TYPE bill_stake AS (
    user_id BIGINT,
    p SMALLINT,
    q SMALLINT
);

-- строчка счёта
CREATE TYPE bill_item AS (
    title TEXT,
    price BIGINT,
    -- произвольный тип. Можем обозначить что-то как чаевые и считать их по особому
    type SMALLINT,
    stakes bill_stake[]
);

-- сколько внёс тот или иной пользователь
CREATE TYPE bill_share AS (
    user_id BIGINT,
    value BIGINT
);

CREATE TABLE accounting_split_the_bill (
    owner BIGINT NOT NULL REFERENCES users(id),
    items bill_item[],
    stakes bill_stake[]
) INHERITS (base_accounting);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
