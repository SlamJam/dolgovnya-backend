package models

import (
	"math/big"
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

func FixInvocesTotal(invoices []Invoice, targetTotal decimal.Decimal) ([]Invoice, error) {
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

	// проверка суммы инвойсов после исправления
	totalInvoicesValue := NewMoney()
	for _, invoice := range invoices {
		totalInvoicesValue.Decimal = totalInvoicesValue.Add(invoice.Value.Decimal)
	}

	if !targetTotal.Equal(totalInvoicesValue.Decimal) {
		return nil, errors.Wrap(ErrInternalAssertion, "fail decimal final test")
	}

	return invoices, nil
}
