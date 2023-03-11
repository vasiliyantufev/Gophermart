package model

import (
	"github.com/vasiliyantufev/gophermart/internal/storage"
	"time"
)

type Order struct {
	ID            int              `json:"id"`
	UserID        int              `json:"user_id"`
	OrderID       int              `json:"order_number"`
	CurrentStatus storage.Statuses `json:"current_status"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}
