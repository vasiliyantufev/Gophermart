package model

import (
	"database/sql"
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

type OrderResponse struct {
	Number     int           `json:"number"`
	Status     string        `json:"status"`
	Accrual    sql.NullInt64 `json:"-"`
	UploadedAt time.Time     `json:"uploaded_at"`
}

type OrderID struct {
	Order   int
	Status  string
	Accrual int
}
