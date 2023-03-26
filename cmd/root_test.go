package cmd_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/SlamJam/dolgovnya-backend/internal/app/models"
	"github.com/SlamJam/dolgovnya-backend/internal/app/services"
	"github.com/SlamJam/dolgovnya-backend/internal/app/storage/pgsql"
	"github.com/SlamJam/dolgovnya-backend/internal/bootstrap/fxapp"
	"github.com/doug-martin/goqu/v9"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
)

func init() {
	goqu.SetDefaultPrepared(true)
}

var psql = goqu.Dialect("postgres")

type ValueExpression interface {
	exp.Expression
	exp.Aliaseable
}

func Values(alias string, firstVal goqu.Vals, restVals ...goqu.Vals) ValueExpression {
	vals := make([]goqu.Vals, 0, 1+len(restVals))
	vals = append(vals, firstVal)
	vals = append(vals, restVals...)

	var args []interface{}
	parts := make([]string, 0, len(vals))
	for _, val := range vals {
		placeholders := make([]string, 0, len(val))
		for _, v := range val {
			args = append(args, v)
			placeholders = append(placeholders, "?")
		}

		parts = append(parts, "("+strings.Join(placeholders, ",")+")")
	}

	return goqu.L("(VALUES "+strings.Join(parts, ",")+") AS "+alias, args...)
}

func TestSQL(t *testing.T) {
	q := psql.Update("table_name").
		From(
			// goqu.L("VALUES (?, ?), (?, ?)", 1, 2, 3, 4).As("V(a,b)"),
			Values("V(a,b)",
				goqu.Vals{1, 2},
				goqu.Vals{3, 4},
			),
		).
		Set(goqu.Record{"foo": goqu.L("V.a")}).
		Where(
			goqu.C("q").Eq(goqu.L("V.b")),
		)

	query, qrgs, err := q.ToSQL()
	_, _, _ = query, qrgs, err
}

func NewMoneyFromInt(amount int64) models.Money {
	return models.Money{models.NewMoney().Add(decimal.NewFromInt(amount))}
}

func populateFromApp(t *testing.T, pointers ...any) error {
	stop, err := fxapp.PopulateFromApp(context.Background(), pointers...)
	if err != nil {
		return nil
	}

	t.Cleanup(func() {
		_ = stop()
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

func TestUserBalances(t *testing.T) {
	require := require.New(t)

	var s *pgsql.Storage
	require.NoError(
		populateFromApp(t, &s),
	)

	userID := models.UserID(5)

	balances, err := s.GetUserBalances(context.Background(), userID)
	require.NoError(err)

	fmt.Println(balances)
}

func TestUserAccount(t *testing.T) {
	require := require.New(t)

	var s *pgsql.Storage
	require.NoError(
		populateFromApp(t, &s),
	)

	userID := models.UserID(1)

	acc, err := s.GetUserAccount(context.Background(), userID)
	require.NoError(err)

	fmt.Println(acc)
}
