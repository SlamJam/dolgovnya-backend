package models

import "github.com/pkg/errors"

type BillPayment struct {
	UserID UserID
	Amount Money
}

func (bp *BillPayment) Validate() error {
	if err := bp.Amount.Validate(); err != nil {
		return errors.Wrapf(err, "Amount precision more than money")
	}

	return nil
}
