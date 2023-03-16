package models

import "fmt"

type UserID int64

func (uid *UserID) String() string {
	return fmt.Sprintf("UserId(%d)", uid)
}

type User struct {
	ID    UserID
	Title string
}
