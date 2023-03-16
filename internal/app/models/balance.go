package models

type NetKind int

const (
	Unknown NetKind = iota
	Zero
	Surplus
	Scarcity
)

type Balance struct {
	Credit MoneyRat
	Debit  MoneyRat
}

func NewBalance() *Balance {
	return &Balance{
		Debit:  NewMoneyRat(),
		Credit: NewMoneyRat(),
	}
}

func (b *Balance) AbsNet() (MoneyRat, NetKind) {
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

func (b *Balance) IsSurplus() (MoneyRat, bool) {
	v, kind := b.AbsNet()
	return v, kind == Surplus
}

func (b *Balance) IsScarcity() (MoneyRat, bool) {
	v, kind := b.AbsNet()
	return v, kind == Scarcity
}

func (b *Balance) IsZero() bool {
	_, kind := b.AbsNet()
	return kind == Zero
}
