package model

import "time"

type Token struct {
	UserID    int       `json:"id"`
	Login     string    `json:"login"`
	Token     string    `json:"token"`
	DeletedAt time.Time `json:"deleted_at"`
}
