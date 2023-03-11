package token

import (
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
)

type TokenRepository interface {
	Created(*model.Token) error
	Validated(*model.Token) error
}

type Token struct {
	token *TokenRepository
	db    *database.DB
}

func New(db *database.DB) *Token {
	return &Token{
		db: db,
	}
}

func (t *Token) Create(token *model.Token) error {

	return nil
}

func (t *Token) Validated(token *model.Token) error {

	return nil
}
