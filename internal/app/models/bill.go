package models

import (
	"fmt"
	"math/big"

	"github.com/pkg/errors"
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
	maxIntFloat64            = float64(1 << 53)
	CurrentBillSchemaVersion = 1
)

type Bill struct {
	ID       BillID        `json:"-"`
	Items    []BillItem    `json:",omitempty"`
	Payments []BillPayment `json:",omitempty"`
}

type BillID int64

func (bid BillID) String() string {
	return fmt.Sprintf("BillID(%d)", bid)
}

func (b *Bill) GetSchemaVersion() int {
	return CurrentBillSchemaVersion
}

func (b *Bill) TotalPayment() Money {
	totalPayment := NewMoney()
	for _, payment := range b.Payments {
		totalPayment.Decimal = totalPayment.Add(payment.Amount.Decimal)
	}

	return totalPayment
}

func (b *Bill) TotalPrice() Money {
	totalPrice := NewMoney()
	for _, item := range b.Items {
		totalPrice.Decimal = totalPrice.Add(item.TotalPrice().Decimal)
	}

	return totalPrice
}

func (b *Bill) Validate() error {
	for index, item := range b.Items {
		if err := item.Validate(); err != nil {
			return errors.Wrapf(err, "item at index %d is invalid", index)
		}
	}

	for index, p := range b.Payments {
		if err := p.Validate(); err != nil {
			return errors.Wrapf(err, "payment at index %d is invalid", index)
		}
	}

	totalPrice := b.TotalPrice()
	if err := totalPrice.Validate(); err != nil {
		return errors.Wrapf(err, "TotalPrice has error")
	}

	totalPayment := b.TotalPayment()
	if err := totalPayment.Validate(); err != nil {
		return errors.Wrapf(ErrMoneyPrecision, "TotalPayment has error")
	}

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
