package cmd_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/SlamJam/dolgovnya-backend/internal/app/models"
	"github.com/SlamJam/dolgovnya-backend/internal/app/services"
	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxapp"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func NewMoneyFromInt(amount int64) models.Money {
	return models.Money{models.NewMoney().Add(decimal.NewFromInt(amount))}
}

func populateFromApp(t *testing.T, pointers ...any) error {
	opts := make([]fx.Option, 0, len(pointers)+1)
	opts = append(opts, fxapp.Module)

	for _, p := range pointers {
		opts = append(opts, fx.Populate(p))
	}

	app := fx.New(opts...)

	if err := app.Start(context.Background()); err != nil {
		return err
	}

	t.Cleanup(func() {
		app.Stop(context.Background())
	})

	return nil
}

func TestCreateUsers(t *testing.T) {
	require := require.New(t)

	var s services.SplitTheBillStorage
	require.NoError(
		populateFromApp(t, &s),
	)

	for _, user := range []string{"vasya1", "vasya2", "vasya3", "vasya4", "vasya5"} {
		userId, err := s.CreateUser(context.Background(), user)
		require.NoError(err)
		fmt.Println(userId)
	}
}

func TestCreateBill(t *testing.T) {
	require := require.New(t)

	var s services.SplitTheBillStorage
	require.NoError(
		populateFromApp(t, &s),
	)

	bill := models.Bill{
		Items: []models.BillItem{
			{
				Title:       "Торт",
				PricePerOne: NewMoneyFromInt(100),
				Quantity:    1,
				Shares: []models.BillShare{
					{UserID: 1, Share: 1},
					{UserID: 2, Share: 1},
					{UserID: 3, Share: 1},
				},
			},
		},
		Payments: []models.BillPayment{
			{UserID: 5, Amount: NewMoneyFromInt(100)},
		},
	}

	require.NoError(
		bill.Validate(),
	)

	invoices, err := bill.ToInvoices()
	require.NoError(err)

	for _, item := range bill.Items {
		fmt.Println(item)
	}
	fmt.Println("===============")
	for _, payment := range bill.Payments {
		fmt.Println(payment)
	}
	fmt.Println("===============")
	for _, invoice := range invoices {
		fmt.Println(invoice)
	}

	userID := models.UserID(1)
	billID, err := s.SaveSplittedBill(context.Background(), userID, bill)
	require.NoError(err)

	fmt.Println(billID)
}
