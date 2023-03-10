package balance

import (
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
)

type BalanceRepository interface {
	GetBalance(*model.User, database.DB) error
	WithDraw(*model.User, database.DB) error
	WithDrawals(*model.User, database.DB) error
}

type Balance struct {
	balance *BalanceRepository
}

func (b *Balance) GetBalance(*model.User, database.DB) error {

	return nil
}

func (b *Balance) WithDraw(*model.User, database.DB) error {

	return nil
}

func (b *Balance) WithDrawals(*model.User, database.DB) error {

	return nil
}
