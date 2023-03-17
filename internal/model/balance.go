package model

import (
	"database/sql"
	"time"
)

type Balance struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	OrderID   int       `json:"order_id"`
	Delta     float64   `json:"delta"`
	CreatedAt time.Time `json:"created_at"`
}

type BalanceUser struct {
	Current   sql.NullFloat64
	Withdrawn sql.NullFloat64
}

type BalanceWithdraw struct {
	Order int
	Sum   float64
}

type BalanceWithdrawals struct {
	Order       int
	Sum         float64
	ProcessedAt time.Time
}
