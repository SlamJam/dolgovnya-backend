package models

import (
	"math/big"
	"sort"

	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

var (
	ErrNoShares     = errors.New("shares are empty")
	ErrZeroQuantity = errors.New("quantity is zero")
)

type BillItem struct {
	Title       string
	PricePerOne Money
	Quantity    uint
	Type        uint8
	Shares      []BillShare
}

type BillShare struct {
	UserID UserID
	Share  uint32
}

func (bi *BillItem) Validate() error {
	if bi.Quantity < 1 {
		return ErrZeroQuantity
	}

	if len(bi.Shares) == 0 {
		return ErrNoShares
	}

	if err := bi.PricePerOne.Validate(); err != nil {
		return errors.Wrap(err, "PricePerOne has error")
	}

	if err := bi.TotalPrice().Validate(); err != nil {
		return errors.Wrap(err, "TotalPrice has error")
	}

	return nil
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

type numbers interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func findGCD[T numbers](nums []T) T {
	sort.Slice(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})

	l, r := nums[0], nums[len(nums)-1]
	for i := l; i >= 0; i-- {
		if l%i == 0 && r%i == 0 {
			return i
		}
	}
	return 1
}

func (bi *BillItem) SimplifyShares() {
	shares := make([]uint32, 0, len(bi.Shares))
	for _, share := range bi.Shares {
		shares = append(shares, share.Share)
	}

	gdc := findGCD(shares)
	if gdc == 1 {
		return
	}

	for _, share := range bi.Shares {
		share.Share = share.Share / gdc
	}
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
