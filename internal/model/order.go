package model

import (
	"github.com/vasiliyantufev/gophermart/internal/storage/statuses"
	"time"
)

type OrderDB struct {
	ID            int               `json:"id"`
	UserID        int               `json:"user_id"`
	OrderID       string            `json:"order_number"`
	CurrentStatus statuses.Statuses `json:"current_status"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

type OrdersResponseGophermart struct {
	Number     string            `json:"number"`
	Status     statuses.Statuses `json:"status"`
	Accrual    *int              `json:"accrual,omitempty"`
	UploadedAt time.Time         `json:"uploaded_at"`
}

type OrderResponseAccrual struct {
	Order   string            `json:"order"`
	Status  statuses.Statuses `json:"status"`
	Accrual int               `json:"accrual"`
}
