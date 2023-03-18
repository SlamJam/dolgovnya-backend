package pgsql

import (
	"context"
	"database/sql/driver"
	"encoding/json"

	"github.com/SlamJam/dolgovnya-backend/internal/app/models"
	"github.com/pkg/errors"
)

type dbBill models.Bill

// Value make the Attrs struct implement the driver.Valuer interface.
func (b dbBill) Value() (driver.Value, error) {
	res, err := json.Marshal(b)
	if err == nil {
		return string(res), err
	}
	return res, err
}

// Scan make the PriceData struct implement the sql.Scanner interface.
func (b *dbBill) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(bytes, b)
}

type dbInvoice struct {
	Accounting int64         `db:"owner_accounting"`
	UserFrom   models.UserID `db:"user_from"`
	UserTo     models.UserID `db:"user_to"`
	Amount     models.Money  `db:"amount"`
}

func (s *Storage) SaveSplittedBill(ctx context.Context, ownerID models.UserID, bill models.Bill) (models.BillID, error) {
	invoices, err := bill.ToInvoices()
	if err != nil {
		return 0, errors.WithStack(err)
	}

	tx, err := s.pool.BeginTxx(ctx, nil)
	if err != nil {
		return 0, errors.WithStack(err)
	}
	defer tx.Rollback()

	var owningObjID int64
	psql.Insert("owner_objects").
		Columns(
			"user_id",
		).
		Values(
			ownerID,
		).
		Suffix(`RETURNING "id"`).
		RunWith(tx).
		QueryRow().
		Scan(&owningObjID)

	if err != nil {
		return 0, errors.WithStack(err)
	}

	var billID models.BillID
	psql.Insert("accounting_split_the_bill").
		Columns(
			"user_id",
			"owning_object_id",
			"schema_version",
			"bill",
		).
		Values(
			ownerID,
			owningObjID,
			bill.SchemaVersion,
			dbBill(bill),
		).
		Suffix(`RETURNING "id"`).
		RunWith(tx).
		QueryRow().
		Scan(&billID)

	if err != nil {
		return 0, errors.WithStack(err)
	}

	q := psql.Insert("accounting_entries").
		Columns(
			"user_id",
			"owning_object_id",
			"user_from",
			"user_to",
			"amount",
		)

	for _, invoice := range invoices {
		q = q.Values(
			ownerID,
			owningObjID,
			invoice.UserFrom,
			invoice.UserTo,
			invoice.Value.Decimal,
		)
	}

	sql, args := q.MustSql()
	_, _ = sql, args

	rows, err := q.RunWith(tx).Query()
	if err != nil {
		return 0, errors.WithStack(err)
	}
	defer rows.Close()

	err = tx.Commit()
	if err != nil {
		return 0, errors.WithStack(err)
	}

	return billID, nil
}

func (s *Storage) ListUserBills(ctx context.Context, userID models.UserID) ([]models.Bill, error) {
	return nil, nil
}

func (s *Storage) GetBills(ctx context.Context, billIDs []models.BillID) ([]models.Bill, error) {
	return nil, nil
}

func (s *Storage) DeleteBills(ctx context.Context, billIDs []models.BillID) ([]models.Bill, error) {
	return nil, nil
}
