package models

import (
	"math/big"
	"sort"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
	"go.uber.org/multierr"
)

var (
	ErrNoShares          = errors.New("shares are empty")
	ErrZeroQuantity      = errors.New("quantity is zero")
	ErrDiscrepancy       = errors.New("total price and total payments must be equal")
	ErrBalanceNotZero    = errors.New("balance isn't zero")
	ErrInternalAssertion = errors.New("internal assertion")
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

type Bill struct {
	ID       BillID        `json:"-"`
	Items    []BillItem    `json:",omitempty"`
	Payments []BillPayment `json:",omitempty"`
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

func (b *Bill) ToInvoices() ([]Invoice, error) {
	if err := b.Validate(); err != nil {
		return nil, err
	}

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

	type userMoneyRat struct {
		UserID UserID
		Amount MoneyRat
	}

	credits := []userMoneyRat{}
	debts := []userMoneyRat{}

	for userID, v := range balances {
		sign := v.Sign()
		abs := MoneyRat{big.NewRat(0, 1).Abs(v.Rat)}

		switch sign {
		case -1:
			debts = append(debts, userMoneyRat{UserID: userID, Amount: abs})
		case 1:
			credits = append(credits, userMoneyRat{UserID: userID, Amount: abs})
		}
	}

	sort.Slice(debts, func(i, j int) bool {
		return debts[i].Amount.Cmp(debts[i].Amount.Rat) == 1
	})

	sort.Slice(credits, func(i, j int) bool {
		return credits[i].Amount.Cmp(credits[i].Amount.Rat) == 1
	})

	totalDebit := NewMoneyRat()
	for _, v := range debts {
		totalDebit.Add(totalDebit.Rat, v.Amount.Rat)
	}

	totalCredit := NewMoneyRat()
	for _, v := range credits {
		totalCredit.Add(totalCredit.Rat, v.Amount.Rat)
	}

	if totalDebit.Cmp(totalCredit.Rat) != 0 {
		return nil, ErrBalanceNotZero
	}

	// debts + credits -> []Invoice
	invoices := []Invoice{}
	zeroRat := big.NewRat(0, 1)

	var debtsClosed int
	for _, credit := range credits {
		for _, debt := range debts[debtsClosed:] {
			if credit.Amount.Cmp(zeroRat) == 0 {
				break
			}

			var debited MoneyRat
			switch credit.Amount.Cmp(debt.Amount.Rat) {
			case -1, 0:
				debited = credit.Amount
			case 1:
				debited = debt.Amount
			}

			decimalAmount := decimal.NewFromBigInt(debited.Num(), 0).
				Div(decimal.NewFromBigInt(debited.Denom(), 0))

			invoices = append(invoices, Invoice{
				UserFrom: credit.UserID,
				UserTo:   debt.UserID,
				Value:    Money{decimalAmount},
			})

			credit.Amount.Sub(credit.Amount.Rat, debited.Rat)
			debt.Amount.Sub(debt.Amount.Rat, debited.Rat)

			if debt.Amount.Cmp(zeroRat) == 0 {
				debtsClosed++
			}
		}
	}

	// исправление неточности Decimal'ов. 100 на троих это 33.33, 33.33 и 33.33,
	// итого не достаёт копейки 0.01. Исправленное: 33.33, 33.33 и 33.34 (33.33 + 0.01)
	invoices, err := fixInvocesTotal(invoices, b.TotalPayment().Decimal)
	if err != nil {
		return nil, errors.Wrap(multierr.Append(ErrInternalAssertion, err), "error at fix total")
	}

	// проверка суммы инвойсов после исправления
	totalInvoicesValue := NewMoney()
	for _, invoice := range invoices {
		totalInvoicesValue.Decimal = totalInvoicesValue.Add(invoice.Value.Decimal)
	}

	d1 := NewMoneyRat().Sub(totalCredit.Rat, totalInvoicesValue.Rat())
	if d1.Cmp(zeroRat) != 0 {
		return nil, errors.Wrap(ErrInternalAssertion, "fail rational final test")
	}

	d2 := b.TotalPayment().Sub(totalInvoicesValue.Decimal)
	if !d2.IsZero() {
		return nil, errors.Wrap(ErrInternalAssertion, "fail decimal final test")
	}

	return invoices, nil
}

func fixInvocesTotal(invoices []Invoice, targetTotal decimal.Decimal) ([]Invoice, error) {
	invoicesTotal := NewMoney()
	for _, invoice := range invoices {
		invoicesTotal.Decimal = invoicesTotal.Add(invoice.Value.Decimal)
	}

	discrepancy := targetTotal.Sub(invoicesTotal.Decimal)

	if discrepancy.IsZero() {
		return invoices, nil
	}

	invoiceCount := len(invoices)
	fixStep := discrepancy.DivRound(decimal.NewFromInt(int64(invoiceCount)), targetTotal.Exponent())
	fixThreshold := decimal.New(int64(fixStep.Sign()), targetTotal.Exponent())

	if discrepancy.Abs().Cmp(fixThreshold.Abs()) == -1 {
		return nil, errors.Wrapf(ErrInternalAssertion, "fix value (:s) is less than (%s)", discrepancy, fixThreshold)
	}

	if fixStep.Sign() == 1 {
		fixStep = decimal.Max(fixStep, fixThreshold)
	} else {
		fixStep = decimal.Min(fixStep, fixThreshold)
	}

	currentFix := decimal.New(0, 0)
	for i := range invoices {
		inv := &invoices[i]

		if i != invoiceCount-1 {
			invoices[i].Value = Money{inv.Value.Add(fixStep)}
			currentFix = currentFix.Add(fixStep)
		} else {
			invoices[i].Value = Money{inv.Value.Add(discrepancy.Sub(currentFix))}
		}
	}

	return invoices, nil
}
