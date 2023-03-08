package model

import "time"

type Order struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	OrderNumber int       `json:"order_number"`
	Status      string    `json:"status"`
	Accrual     int       `json:"accrual"`
	UploadedAt  time.Time `json:"uploaded_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
