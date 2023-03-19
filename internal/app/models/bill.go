package models

import (
	"fmt"
	"math/big"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/multierr"
)

var (
	ErrNoShares              = errors.New("shares are empty")
	ErrZeroQuantity          = errors.New("quantity is zero")
	ErrDiscrepancy           = errors.New("total price and total payments must be equal")
	ErrInternalAssertion     = errors.New("internal assertion")
	ErrTotalDecimalPrecision = errors.Wrap(ErrInternalAssertion, "decimal precision of total")
)

const (
	maxIntFloat64     = float64(1 << 53)
	BillSchemaVersion = 1
)

type BillShare struct {
	UserID UserID
	Share  uint32
}

type BillPayment struct {
	UserID UserID
	Amount Money
}

type BillItem struct {
	Title       string
	PricePerOne Money
	Quantity    uint
	Type        uint8
	Shares      []BillShare
}

func (bi *BillItem) TotalPrice() Money {
	return Money{bi.PricePerOne.Mul(decimal.NewFromInt(int64(bi.Quantity)))}
}

func (bi *BillItem) TotalShare() int64 {
	var totalShare int64
	for _, share := range bi.Shares {
		totalShare += int64(share.Share)
	}

	return totalShare
}

func (bi *BillItem) SharePricesByUser() map[UserID]MoneyRat {
	res := map[UserID]MoneyRat{}
	price := bi.TotalPrice().Rat()
	totalShare := bi.TotalShare()

	for _, share := range bi.Shares {
		ratio := big.NewRat(int64(share.Share), totalShare)
		sharePrice := MoneyRat{NewMoneyRat().Mul(price, ratio)}

		old := NewMoneyRat()
		if v, ok := res[share.UserID]; ok {
			old = v
		}

		res[share.UserID] = MoneyRat{old.Add(old.Rat, sharePrice.Rat)}
	}

	return res
}

type BillID int64

func (bid BillID) String() string {
	return fmt.Sprintf("BillID(%d)", bid)
}

type Bill struct {
	ID       BillID        `json:"-"`
	Items    []BillItem    `json:",omitempty"`
	Payments []BillPayment `json:",omitempty"`
}

func (b *Bill) GetSchemaVersion() int {
	return BillSchemaVersion
}

func (b *Bill) TotalPayment() Money {
	totalPayment := NewMoney()
	for _, payment := range b.Payments {
		totalPayment.Decimal = totalPayment.Add(payment.Amount.Decimal)
	}

	return totalPayment
}

func (b *Bill) Validate() error {
	var totalPrice Money
	for itemIndex, item := range b.Items {
		switch {
		case item.Quantity < 1:
			return errors.Wrapf(ErrZeroQuantity, "item at index %d has no quantity", itemIndex)

		case len(item.Shares) == 0:
			return errors.Wrapf(ErrNoShares, "item at index %d has no shares", itemIndex)
		}

		totalPrice.Decimal = totalPrice.Add(item.TotalPrice().Decimal)
	}

	totalPayment := b.TotalPayment()
	if !totalPayment.Equal(totalPrice.Decimal) {
		return errors.Wrapf(ErrDiscrepancy, "total price: %s, total payments: %s", totalPrice, totalPayment)
	}

	return nil
}

func (b *Bill) DebitByUser() map[UserID]MoneyRat {
	res := map[UserID]MoneyRat{}

	for _, item := range b.Items {
		for userId, moneyRat := range item.SharePricesByUser() {
			old := NewMoneyRat()
			if v, ok := res[userId]; ok {
				old = v
			}
			res[userId] = MoneyRat{old.Add(old.Rat, moneyRat.Rat)}
		}
	}

	return res
}

func (b *Bill) CreditByUser() map[UserID]MoneyRat {
	res := map[UserID]MoneyRat{}

	for _, payment := range b.Payments {
		old := NewMoneyRat()
		if v, ok := res[payment.UserID]; ok {
			old = v
		}

		res[payment.UserID] = MoneyRat{old.Add(old.Rat, payment.Amount.Rat())}
	}

	return res
}

func (b *Bill) BalanceByUser() map[UserID]MoneyRat {
	balances := map[UserID]MoneyRat{}

	for userID, value := range b.DebitByUser() {
		v, ok := balances[userID]
		if !ok {
			v = NewMoneyRat()
		}

		v.Sub(v.Rat, value.Rat)
		balances[userID] = v
	}

	for userID, value := range b.CreditByUser() {
		v, ok := balances[userID]
		if !ok {
			v = NewMoneyRat()
		}

		v.Add(v.Rat, value.Rat)
		balances[userID] = v
	}

	return balances
}

func (b *Bill) ToInvoices() ([]Invoice, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}

	zeroRat := big.NewRat(0, 1)
	balancesMap := b.BalanceByUser()

	invoices, err := InvoicesFromBalances(balancesMap)
	if err != nil {
		return nil, errors.Wrap(err, "fail construct Invoices")
	}

	// Исправление неточности конечной суммы при переходе от Rational к Decimal.
	// Напримео, 100 на троих это 33.33, 33.33 и 33.33, итого не достаёт копейки 0.01.
	// Исправленное: 33.33, 33.33 и 33.34 (33.33 + 0.01).

	// Считаем сумму выданных денег. Т.к. баланс нулевой, то это равно и сумму одолженых.
	totalCredit := NewMoneyRat()
	for _, v := range balancesMap {
		if v.Rat.Cmp(zeroRat) == 1 {
			totalCredit.Add(totalCredit.Rat, v.Rat)
		}
	}

	// в округлении накидываем в сторону должников
	totalCreditDecimal := NewMoneyFromBig(totalCredit.Num()).
		Div(NewMoneyFromBig(totalCredit.Denom()).Decimal).
		RoundCeil(MoneyPrecision).
		Truncate(MoneyPrecision)

	invoices, err = FixInvocesTotal(invoices, totalCreditDecimal)
	if err != nil {
		return nil, errors.Wrap(multierr.Append(ErrInternalAssertion, err), "error at fix total")
	}

	return invoices, nil
}
