package pgsql

import (
	"context"

	"github.com/SlamJam/dolgovnya-backend/internal/app/models"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

func (s *Storage) GetBalanceForUser(ctx context.Context, userID models.UserID) (models.Account, error) {
	var balance models.Account

	tx, err := s.pool.BeginTxx(ctx, nil)
	if err != nil {
		return balance, errors.WithStack(err)
	}
	defer tx.Rollback()

	grp, _ := errgroup.WithContext(ctx)
	grp.Go(func() error {
		return psql.Select("sum(amount) as debit").
			From("accounting_entries").
			Where("user_from = ?", userID).
			RunWith(tx).
			QueryRow().
			Scan(&balance.Debit)
	})
	grp.Go(func() error {
		return psql.Select("sum(amount) as credit").
			From("accounting_entries").
			Where("user_to = ?", userID).
			RunWith(tx).
			QueryRow().
			Scan(&balance.Credit)
	})

	err = grp.Wait()

	return balance, errors.WithStack(err)
}
