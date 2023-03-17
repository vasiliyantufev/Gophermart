package balance

import (
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
	"github.com/vasiliyantufev/gophermart/internal/storage/errors"
	"time"
)

type Balancer interface {
	GetBalance(userId int) (*model.BalanceUser, error)
	Accrue(userId int, orderID *model.OrderID) error
	WithDraw(userId int, wd *model.BalanceWithdraw) error
	WithDrawals(userId int) ([]model.BalanceWithdrawals, error)
}

type Balance struct {
	db *database.DB
}

func New(db *database.DB) *Balance {
	return &Balance{
		db: db,
	}
}

func (b *Balance) GetBalance(userId int) (*model.BalanceUser, error) {

	balanceUser := &model.BalanceUser{}

	if err := b.db.Pool.QueryRow("select sum(delta) as current, sum(case when delta < 0 then delta end) as withdrawn "+
		"from balance where user_id = $1", userId).Scan(
		&balanceUser.Current,
		&balanceUser.Withdrawn,
	); err != nil {
		return nil, err
	}
	return balanceUser, nil
}

func (b *Balance) Accrue(userId int, orderID *model.OrderID) error {

	var id int
	return b.db.Pool.QueryRow(
		"INSERT INTO balance (user_id, order_id, delta, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		userId,
		orderID.Order,
		orderID.Accrual,
		time.Now(),
	).Scan(&id)
}

func (b *Balance) WithDraw(userId int, wd *model.BalanceWithdraw) error {

	var sum float64
	err := b.db.Pool.QueryRow("select sum(delta) as balance from balance where user_id = $1", userId).Scan(
		&sum,
	)
	if err != nil {
		return err
	}

	if sum < wd.Sum {
		return errors.ErrNotFunds
	}

	var id int
	return b.db.Pool.QueryRow(
		"INSERT INTO balance (user_id, order_id, delta, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		userId,
		wd.Order,
		wd.Sum,
		time.Now(),
	).Scan(&id)
}

func (b *Balance) WithDrawals(userId int) ([]model.BalanceWithdrawals, error) {

	var withdraw model.BalanceWithdrawals
	var withdrawals []model.BalanceWithdrawals

	query := "SELECT order_id, delta, created_at FROM balance " +
		"WHERE delta < 0 and user_id = $1 ORDER BY created_at"

	rows, err := b.db.Pool.Query(query, userId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err = rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdraw)
	}

	return withdrawals, nil
}
