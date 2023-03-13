package balance

import (
	"errors"
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
	"time"
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

func (b *Balance) GetBalance(user *model.User) (*model.BalanceUser, error) {

	// сумма баллов лояльности за весь период регистрации баллов.
	// сумма использованных.

	balanceUser := &model.BalanceUser{}

	if err := b.db.Pool.QueryRow("select sum(delta) as current, sum(case when delta < 0 then delta end) as withdrawn "+
		"from balance where user_id = $1", user.ID).Scan(
		&balanceUser.Current,
		&balanceUser.Withdrawn,
	); err != nil {
		return nil, err
	}
	return balanceUser, nil
}

func (b *Balance) WithDraw(user *model.User, wd *model.BalanceWithdraw) error {

	var sum float64
	err := b.db.Pool.QueryRow("select sum(delta) as balance from user where user_id = $1", user.ID).Scan(
		&sum,
	)
	if err != nil {
		return err
	}

	balance := &model.Balance{}

	if sum > wd.Sum {
		return b.db.Pool.QueryRow(
			"INSERT INTO balance (user_id, order_id, delta, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
			user.ID,
			wd.Order,
			wd.Sum,
			time.Now(),
		).Scan(&balance.ID)
	}

	return errors.New("There are not enough funds on the account")
}

func (b *Balance) WithDrawals(user *model.User) ([]model.BalanceWithdrawals, error) {

	var withdraw model.BalanceWithdrawals
	var withdrawals []model.BalanceWithdrawals

	query := "SELECT order_id, delta, created_at FROM balance WHERE delta < 0 and user_id = $1 ORDER BY created_at"

	rows, err := b.db.Pool.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err = rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.Processed_at); err != nil {
			return nil, err
		}

		withdrawals = append(withdrawals, withdraw)
	}

	return withdrawals, nil
}
