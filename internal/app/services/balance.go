package services

import (
	"context"

	"github.com/SlamJam/dolgovnya-backend/internal/app/logger"
	"github.com/SlamJam/dolgovnya-backend/internal/app/models"
)

type BalanceStorage interface {
	// GetInvoices(context.Context, models.UserID) ([]models.Invoice, error)
	GetBalanceForUser(context.Context, models.UserID) (models.Account, error)
}

type BalanceService struct {
	storage BalanceStorage
	logger  logger.Logger
}

func NewBalanceService(storage BalanceStorage, log logger.Logger) *BalanceService {
	return &BalanceService{
		storage: storage,
		logger:  log,
	}
}

func (s *BalanceService) log(ctx context.Context) logger.Logger {
	return logger.FromCtxOrDefault(ctx, s.logger)
}

func (s *BalanceService) GetBalance(ctx context.Context, userID models.UserID) (models.Account, error) {
	acc, err := s.storage.GetBalanceForUser(ctx, userID)
	if err != nil {
		s.log(ctx).Error().Err(err).
			Int64("user_id", int64(userID)).
			Msg("fail to get user's balance from storage")
	}

	return acc, err
}
