package services

import (
	"context"

	"github.com/SlamJam/dolgovnya-backend/internal/app/models"
)

type SplitTheBillStorage interface {
	// User
	CreateUser(context.Context, string) (models.UserID, error)

	// Bill
	SaveSplittedBill(context.Context, models.UserID, models.Bill) (models.BillID, error)
	ListUserBills(context.Context, models.UserID) ([]models.Bill, error)
	GetBills(context.Context, []models.BillID) ([]models.Bill, error)
	DeleteBills(context.Context, []models.BillID) ([]models.Bill, error)
}
