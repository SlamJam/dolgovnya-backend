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
	q := psql.Select("user_id, sum(amount)").
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
			)`, userID, userID)
		// RunWith(s.pool).
		// QueryRow().
		// Scan(&balance.Debit, &balance.Credit)

	sql, args, err := q.ToSql()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	rows, err := s.pool.Queryx(sql, args...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()

	result := map[models.UserID]models.Money{}
	for rows.Next() {
		var userID models.UserID
		var amount models.Money

		err = rows.Scan(&userID, &amount)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		if _, ok := result[userID]; ok {
			return nil, errors.New("user is not uniq. Query broken")
		}

		result[userID] = amount
	}

	return result, nil
}
