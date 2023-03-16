package model

import "time"

type Token struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	DeletedAt time.Time `json:"deleted_at"`
}

type TokenUser struct {
	UserID    int       `json:"id"`
	Login     string    `json:"login"`
	Token     string    `json:"token"`
	DeletedAt time.Time `json:"deleted_at"`
}
