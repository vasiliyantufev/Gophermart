package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Login     string    `json:"login"`
	Password  string    `json:"password,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type UserPayload struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
}
