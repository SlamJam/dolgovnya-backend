package models

import (
	"math/big"
	"math/rand"
	"sort"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	ErrBalanceNotZero = errors.New("balance isn't zero")
)

func InvoicesFromBalances(balancesMap map[UserID]MoneyRat) ([]Invoice, error) {
	zeroRat := big.NewRat(0, 1)

	type userMoneyRat struct {
		UserID UserID
		Amount MoneyRat
	}

	total := NewMoneyRat()
	balances := make([]userMoneyRat, 0, len(balancesMap))
	for userID, v := range balancesMap {
		// Если заплатил за себя, то ни ты, ни тебе никто не должен. Пропускаем.
		if v.Cmp(zeroRat) == 0 {
			continue
		}

		balances = append(balances, userMoneyRat{UserID: userID, Amount: v.Copy()})
		total.Add(total.Rat, v.Rat)
	}

	if total.Cmp(zeroRat) != 0 {
		return nil, ErrBalanceNotZero
	}

	sort.Slice(balances, func(i, j int) bool {
		return balances[i].Amount.Cmp(balances[j].Amount.Rat) == 1
	})

	// balances -> []Invoice
	invoices := []Invoice{}
	var i, j int = 0, len(balances) - 1
	for i < j {
		credit := balances[i]
		debt := balances[j]

		if credit.Amount.Cmp(zeroRat) != 1 {
			return nil, errors.Wrap(ErrInternalAssertion, "credit is not positive")
		}

		if debt.Amount.Cmp(zeroRat) != -1 {
			return nil, errors.Wrap(ErrInternalAssertion, "debt is not negative")
		}

		creditAbs := NewMoneyRat()
		creditAbs.Abs(credit.Amount.Rat)
		debtAbs := NewMoneyRat()
		debtAbs.Abs(debt.Amount.Rat)

		var debited MoneyRat
		switch creditAbs.Cmp(debtAbs.Rat) {
		case -1, 0:
			debited = creditAbs
		case 1:
			debited = debtAbs
		}

		invoices = append(invoices, Invoice{
			UserFrom: credit.UserID,
			UserTo:   debt.UserID,
			Value:    debited.Money(),
		})

		credit.Amount.Sub(credit.Amount.Rat, debited.Rat)
		debt.Amount.Add(debt.Amount.Rat, debited.Rat)

		switch credit.Amount.Cmp(zeroRat) {
		case -1:
			return nil, errors.Wrap(ErrInternalAssertion, "credit overdraft")
		// переходим к следующему кредитному балансу
		case 0:
			i++
		// остался остаток, продолдаем
		case 1:
		}

		switch debt.Amount.Cmp(zeroRat) {
		// остался долг, продолдаем
		case -1:
		// переходим к следующему дебитному балансу
		case 0:
			j--
		case 1:
			return nil, errors.Wrap(ErrInternalAssertion, "debt overdraft")
		}
	}

	return invoices, nil
}

func InvoicesTotal(invoices []Invoice) Money {
	total := NewMoney()
	for _, invoice := range invoices {
		total.Decimal = total.Add(invoice.Value.Decimal)
	}

	return total
}

func FixInvocesTotal(invoices []Invoice, targetTotal Money) ([]Invoice, error) {
	if err := targetTotal.Validate(); err != nil {
		return nil, errors.Wrap(err, "target total has error")
	}

	invoicesTotal := InvoicesTotal(invoices)
	discrepancy := targetTotal.Sub(invoicesTotal.Decimal)

	if discrepancy.IsZero() {
		return invoices, nil
	}

	invoiceCount := len(invoices)
	fixStep := discrepancy.DivRound(decimal.NewFromInt(int64(invoiceCount)), targetTotal.Exponent())
	minStep := decimal.New(int64(fixStep.Sign()), MoneyPrecision)

	if discrepancy.Abs().Cmp(minStep.Abs()) == -1 {
		return nil, errors.Wrapf(ErrInternalAssertion, "fix value (:s) is less than (%s)", discrepancy, minStep)
	}

	fixStep = decimal.Max(fixStep.Abs(), minStep.Abs()).
		Mul(decimal.NewFromInt(int64(fixStep.Sign())))

	remainderFix := discrepancy.Copy()
	for i := 0; i < invoiceCount && !remainderFix.IsZero(); i++ {
		invoices[i].Value = Money{invoices[i].Value.Add(fixStep)}
		remainderFix = remainderFix.Sub(fixStep)
	}

	if !remainderFix.IsZero() {
		index := rand.Int() % invoiceCount
		invoices[index].Value = Money{invoices[index].Value.Add(remainderFix)}
	}

	// проверка суммы инвойсов после исправления
	if !targetTotal.Equal(InvoicesTotal(invoices).Decimal) {
		return nil, errors.Wrap(ErrInternalAssertion, "fail decimal final test")
	}

	filtered := make([]Invoice, 0, len(invoices))
	for _, inv := range invoices {
		if !inv.Value.IsZero() {
			filtered = append(filtered, inv)
		}
	}

	return filtered, nil
}
