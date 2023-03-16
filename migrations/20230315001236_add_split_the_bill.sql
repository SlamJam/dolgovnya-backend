-- История разбития счёта --

-- +goose Up
-- +goose StatementBegin

-- какую p/q часть потребил пользователь user_id
CREATE TYPE bill_share AS (
    user_id BIGINT,
    share SMALLINT
);

-- строчка счёта
CREATE TYPE bill_item AS (
    title TEXT,
    price_per_one DECIMAL(15,2),
    quantity    INTEGER,
    -- произвольный тип. Можем обозначить что-то как чаевые и считать их по особому
    type SMALLINT,
    shares bill_share[]
);

-- сколько внёс тот или иной пользователь
CREATE TYPE bill_payment AS (
    user_id BIGINT,
    amount BIGINT CHECK (amount > 0)
);

CREATE TABLE accounting_split_the_bill (
    owner BIGINT NOT NULL REFERENCES users(id),
    items bill_item[],
    payments bill_payment[]
) INHERITS (base_accounting);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
