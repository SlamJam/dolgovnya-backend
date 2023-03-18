package pgsql

import (
	"context"

	"github.com/SlamJam/dolgovnya-backend/internal/app/models"
	"github.com/pkg/errors"
)

func (s *Storage) CreateUser(ctx context.Context, title string) (models.UserID, error) {
	row := psql.Insert("users").
		Columns("title").
		Values(title).
		Suffix(`RETURNING "id"`).
		RunWith(s.pool).QueryRow()

	var userID models.UserID

	err := row.Scan(&userID)
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return userID, nil
}
