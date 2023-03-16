package models

type Invoice struct {
	UserFrom UserID
	UserTo   UserID
	Value    Money
}
