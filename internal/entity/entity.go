package entity

import (
	"time"
)

type User struct {
	ID        int
	Login     string `json:"login"`
	Password  string `json:"password"`
	Cookie    string
	Balance   float64
	Withdrawn float64
}

type Order struct {
	UserID    int       `json:"-"`
	Number    int       `json:"number,string"`
	Status    string    `json:"status"`
	Accrual   float64   `json:"accrual,omitempty"`
	EventTime time.Time `json:"uploaded_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdraw struct {
	Order int       `json:"order,string"`
	Sum   float64   `json:"sum"`
	Time  EventTime `json:"processed_at"`
}

type EventTime time.Time
type CtxUserKey struct{}

func (t EventTime) MarshalJSON() ([]byte, error) {
	newTime := time.Time(t)
	return newTime.MarshalJSON()
}

const (
	SearchByID    = "id"
	SearchByLogin = "login"
)
