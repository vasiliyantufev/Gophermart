package model

import (
	"time"
)

type BalanceWithdraw struct {
	Order int
	Sum   float64
}

type BalanceUserResponse struct {
	Current   *float64 `json:"current,omitempty"`
	Withdrawn *float64 `json:"withdrawn,omitempty"`
}

type BalanceWithdrawalsResponse struct {
	Order       int       `json:"id"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
