package model

import "time"

type BalanceWithdraw struct {
	Order string
	Sum   float64
}

type BalanceUserResponse struct {
	Current   *float64 `json:"current,omitempty"`
	Withdrawn *float64 `json:"withdrawn,omitempty"`
}

type BalanceWithdrawalsResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
