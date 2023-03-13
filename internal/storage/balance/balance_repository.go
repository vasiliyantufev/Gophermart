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
	db      *database.DB
}

func New(db *database.DB) *Balance {
	return &Balance{
		db: db,
	}
}

func (b *Balance) GetBalance(*model.User) error {

	// сумма баллов лояльности за весь период регистрации баллов. - SELECT sum(delta) as balance FROM balance where user_id = $1
	// сумма использованных

	//select
	//	sum(case when val >= 0 then val end) as positive,
	//	sum(case when val < 0 then val end) as negative
	//	from the_data;

	return nil
}

func (b *Balance) WithDraw(*model.User) error {

	return nil
}

func (b *Balance) WithDrawals(*model.User) error {

	return nil
}
