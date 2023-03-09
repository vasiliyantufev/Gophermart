package model

import "time"

type Balance struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Debit     float64   `json:"debit"`
	Credit    float64   `json:"credit"`
	CreatedAt time.Time `json:"created_at"`
}
