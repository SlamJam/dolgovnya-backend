package models

import (
	"math/big"

	"github.com/shopspring/decimal"
)

const (
	MoneyPrecision = 2
)

// Деньги для БД. Конечный результат цепочки вычислений. То, сколько нужно пересести.
// Рациональное число может быть не выразимо в десятичной системе счисления.
// Если требуется отдать 100/3 рублей, то в decimal это 33.33 и 0.00(3) - невязка.
type Money struct{ decimal.Decimal }

func NewMoney() Money {
	return Money{decimal.New(0, -MoneyPrecision)}
}

func NewMoneyFromBig(v *big.Int) Money {
	return Money{NewMoney().Add(decimal.NewFromBigInt(v, 0))}
}

// Тип для вычисления с деньгами. Абсолютная точность в операциях.
// https://github.com/nethruster/go-fraction
type MoneyRat struct{ *big.Rat }

func NewMoneyRat() MoneyRat {
	return MoneyRat{big.NewRat(0, 1)}
}

func (m *MoneyRat) Copy() MoneyRat {
	ret := NewMoneyRat()
	ret.Set(m.Rat)
	return ret
}

func (m *MoneyRat) Money() Money {
	ret := NewMoneyFromBig(m.Num()).
		DivRound(
			NewMoneyFromBig(m.Denom()).Decimal,
			MoneyPrecision,
		)

	return Money{ret}
}
