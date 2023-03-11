package model

import "time"

type Balance struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	OrderID   int       `json:"order_id"`
	Delta     float64   `json:"delta"`
	CreatedAt time.Time `json:"created_at"`
}
