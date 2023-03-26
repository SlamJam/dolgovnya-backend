package pgsql

import (
	"context"

	"github.com/SlamJam/dolgovnya-backend/internal/app/models"
	"github.com/pkg/errors"
)

func (s *Storage) GetUserAccount(ctx context.Context, userID models.UserID) (models.Account, error) {
	var balance models.Account

	err := psql.Select("debit.amount, credit.amount").
		From("debit, credit").
		Prefix(`
			WITH debit as (
				SELECT
					COALESCE(sum(amount), 0) as amount
				FROM accounting_entries
				WHERE user_to = ?
			), credit as (
				SELECT
					COALESCE(sum(amount), 0) as amount
				FROM accounting_entries
				WHERE user_from = ?
			)`, userID, userID).
		RunWith(s.pool).
		QueryRow().
		Scan(&balance.Debit, &balance.Credit)

	return balance, errors.WithStack(err)
}

func (s *Storage) GetUserBalances(ctx context.Context, userID models.UserID) (map[models.UserID]models.Money, error) {
	rows, err := psql.
		Select("user_id, COALESCE(sum(amount), 0)").
		From("balances").
		GroupBy("user_id").
		Prefix(`
			WITH balances as (
				SELECT
					user_to as user_id,
					amount
				FROM accounting_entries
				WHERE user_from = ?

				UNION ALL

				SELECT
					user_from as user_id,
					- amount
				FROM accounting_entries
				WHERE user_to = ?
			)`, userID, userID).
		RunWith(s.pool).
		Query()

	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	return scanToMap[models.UserID, models.Money](rows)
}
