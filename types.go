package api

import (
	"math/rand"
	"time"
)

type CreateAccountRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type Account struct {
	Id          int       `json:"id"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Number      uint64    `json:"bankNumber"`
	Balance     float64   `json:"balance"`
	GoldMemeber bool      `json:"goldMemeber"`
	CreatedAt   time.Time `json:"createdAt"`
}

type TranferRequest struct {
	ToAccount           int     `json:"toAccount"`
	Amount              float64 `json:"amount"`
	AccountBalanceAfter float64 `json:"accountBalanceAfter"`
}

func NewAccount(firstName, lastName string) *Account {
	return &Account{
		FirstName: firstName,
		LastName:  lastName,
		Number:    rand.Uint64(),
		CreatedAt: time.Now().UTC(),
	}
}
