package models

type NetKind int

const (
	Unknown NetKind = iota
	Zero
	Surplus
	Scarcity
)

type Account struct {
	Credit Money
	Debit  Money
}

func NewBalance() *Account {
	return &Account{
		Debit:  NewMoney(),
		Credit: NewMoney(),
	}
}

func (b *Account) AbsNet() (Money, NetKind) {
	var kind NetKind

	diff := b.Credit.Sub(b.Debit.Decimal)
	switch diff.Sign() {
	case -1:
		kind = Scarcity
	case 0:
		kind = Zero
	case 1:
		kind = Surplus
	}

	return Money{diff.Abs()}, kind
}

func (b *Account) IsSurplus() (Money, bool) {
	v, kind := b.AbsNet()
	return v, kind == Surplus
}

func (b *Account) IsScarcity() (Money, bool) {
	v, kind := b.AbsNet()
	return v, kind == Scarcity
}

func (b *Account) IsZero() bool {
	_, kind := b.AbsNet()
	return kind == Zero
}
