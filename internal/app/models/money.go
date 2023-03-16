package models

import (
	"math/big"

	"github.com/shopspring/decimal"
)

// Деньги для БД. Конечный результат цепочки вычислений. То, сколько нужно пересести.
// Рациональное число может быть не выразимо в десятичной системе счисления.
// Если требуется отдать 100/3 рублей, то в decimal это 33.33 и 0.00(3) - невязка.
type Money struct{ decimal.Decimal }

func NewMoney() Money {
	return Money{decimal.New(0, -2)}
}

// Тип для вычисления с деньгами. Абсолютная точность в операциях.
// https://github.com/nethruster/go-fraction
type MoneyRat struct{ *big.Rat }

func NewMoneyRat() MoneyRat {
	return MoneyRat{big.NewRat(0, 1)}
}
