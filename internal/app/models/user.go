package models

import "fmt"

type UserID int64

func (uid UserID) String() string {
	return fmt.Sprintf("UserID(%d)", uid)
}

type User struct {
	ID    UserID
	Title string
}
