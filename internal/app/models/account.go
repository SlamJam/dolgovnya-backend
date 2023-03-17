package models

type NetKind int

const (
	Unknown NetKind = iota
	Zero
	Surplus
	Scarcity
)

type Account struct {
	Credit MoneyRat
	Debit  MoneyRat
}

func NewBalance() *Account {
	return &Account{
		Debit:  NewMoneyRat(),
		Credit: NewMoneyRat(),
	}
}

func (b *Account) AbsNet() (MoneyRat, NetKind) {
	var kind NetKind

	diff := NewMoneyRat().Sub(b.Credit.Rat, b.Debit.Rat)
	switch diff.Sign() {
	case -1:
		kind = Scarcity
	case 0:
		kind = Zero
	case 1:
		kind = Surplus
	}

	return MoneyRat{NewMoneyRat().Abs(diff)}, kind
}

func (b *Account) IsSurplus() (MoneyRat, bool) {
	v, kind := b.AbsNet()
	return v, kind == Surplus
}

func (b *Account) IsScarcity() (MoneyRat, bool) {
	v, kind := b.AbsNet()
	return v, kind == Scarcity
}

func (b *Account) IsZero() bool {
	_, kind := b.AbsNet()
	return kind == Zero
}
