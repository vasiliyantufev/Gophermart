package model

import (
	"time"
)

type Order struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	OrderID       int       `json:"order_number"`
	CurrentStatus string    `json:"current_status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Orders struct {
	Number    int
	Status    string
	Accrual   float64
	UpdatedAt time.Time
}

type OrderID struct {
	Order   int
	Status  string
	Accrual float64
}
