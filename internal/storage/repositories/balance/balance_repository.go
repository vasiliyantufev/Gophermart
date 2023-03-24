package balance

import (
	"time"

	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
	"github.com/vasiliyantufev/gophermart/internal/storage/errors"
)

type Balancer interface {
	GetBalance(userID int) (*model.BalanceUserResponse, error)
	Accrue(userID int, orderID *model.OrderResponseAccrual) error
	CheckBalance(userID int, withdraw *model.BalanceWithdraw) error
	WithDraw(userID int, withdraw *model.BalanceWithdraw) error
	WithDrawals(userID int) ([]model.BalanceWithdrawalsResponse, error)
}

type Balance struct {
	db *database.DB
}

func New(db *database.DB) *Balance {
	return &Balance{
		db: db,
	}
}

func (b *Balance) GetBalance(userID int) (*model.BalanceUserResponse, error) {

	balanceUser := &model.BalanceUserResponse{}

	if err := b.db.Pool.QueryRow("select (sum(accrue) - sum(withdraw)) as current, sum(withdraw) as withdrawn "+
		"from balance where user_id = $1", userID).Scan(
		&balanceUser.Current,
		&balanceUser.Withdrawn,
	); err != nil {
		return nil, err
	}

	return balanceUser, nil
}

func (b *Balance) Accrue(userID int, accrualRequest model.OrderResponseAccrual) error {

	var id int
	return b.db.Pool.QueryRow(
		"INSERT INTO balance (user_id, order_id, accrue, withdraw, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		userID,
		accrualRequest.Order,
		accrualRequest.Accrual,
		0,
		time.Now(),
	).Scan(&id)
}

func (b *Balance) CheckBalance(userID int, withdrawRequest *model.BalanceWithdraw) error {
	var balance *float64
	err := b.db.Pool.QueryRow("select (sum(accrue) - sum(withdraw)) as balance  from balance where user_id = $1", userID).Scan(
		&balance,
	)
	if balance == nil {
		return errors.ErrNotBalance
	}
	if err != nil {
		return err
	}
	if *balance < withdrawRequest.Sum {
		return errors.ErrNotFunds
	}
	return nil
}

func (b *Balance) WithDraw(userID int, withdrawRequest *model.BalanceWithdraw) error {
	var id int
	return b.db.Pool.QueryRow(
		"INSERT INTO balance (user_id, order_id, accrue, withdraw, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		userID,
		withdrawRequest.Order,
		0,
		withdrawRequest.Sum,
		time.Now(),
	).Scan(&id)
}

func (b *Balance) WithDrawals(userID int) ([]model.BalanceWithdrawalsResponse, error) {

	var withdraw model.BalanceWithdrawalsResponse
	var withdrawals []model.BalanceWithdrawalsResponse

	query := "SELECT order_id, withdraw, created_at FROM balance " +
		"WHERE withdraw > 0 and user_id = $1 ORDER BY created_at"

	rows, err := b.db.Pool.Query(query, userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err = rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.ProcessedAt); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdraw)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return withdrawals, nil
}
