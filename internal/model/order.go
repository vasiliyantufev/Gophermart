package model

import (
	"database/sql"
	"github.com/vasiliyantufev/gophermart/internal/storage/statuses"
	"time"
)

type OrderDB struct {
	ID            int               `json:"id"`
	UserID        int               `json:"user_id"`
	OrderID       int               `json:"order_number"`
	CurrentStatus statuses.Statuses `json:"current_status"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type OrdersResponseGophermart struct {
	Number     int               `json:"number"`
	Status     statuses.Statuses `json:"status"`
	Accrual    sql.NullInt64     `json:"accrual"`
	UploadedAt time.Time         `json:"uploaded_at"`
}

type OrderResponseAccrual struct {
	Order   int               `json:"order"`
	Status  statuses.Statuses `json:"status"`
	Accrual int               `json:"accrual"`
}
