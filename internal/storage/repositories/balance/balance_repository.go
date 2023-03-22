package balance

import (
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
	"github.com/vasiliyantufev/gophermart/internal/storage/errors"
	"time"
)

type Balancer interface {
	GetBalance(userId int) (*model.BalanceUserResponse, error)
	Accrue(userId int, orderID *model.OrderResponseAccrual) error
	CheckBalance(userId int, withdraw *model.BalanceWithdraw) error
	WithDraw(userId int, withdraw *model.BalanceWithdraw) error
	WithDrawals(userId int) ([]model.BalanceWithdrawalsResponse, error)
}

type Balance struct {
	db *database.DB
}

func New(db *database.DB) *Balance {
	return &Balance{
		db: db,
	}
}

func (b *Balance) GetBalance(userId int) (*model.BalanceUserResponse, error) {

	balanceUser := &model.BalanceUserResponse{}

	if err := b.db.Pool.QueryRow("select (sum(accrue) - sum(withdraw)) as current, sum(withdraw) as withdrawn "+
		"from balance where user_id = $1", userId).Scan(
		&balanceUser.Current,
		&balanceUser.Withdrawn,
	); err != nil {
		return nil, err
	}

	//current := fmt.Sprintf("%.2f", *balanceUser.Current)
	//*balanceUser.Current, _ = strconv.ParseFloat(current, 64)
	//withdrawn := fmt.Sprintf("%.2f", *balanceUser.Withdrawn)
	//*balanceUser.Withdrawn, _ = strconv.ParseFloat(withdrawn, 64)
	//*balanceUser.Withdrawn *= -1

	return balanceUser, nil
}

func (b *Balance) Accrue(userId int, accrualRequest model.OrderResponseAccrual) error {

	var id int
	return b.db.Pool.QueryRow(
		"INSERT INTO balance (user_id, order_id, accrue, withdraw, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		userId,
		accrualRequest.Order,
		accrualRequest.Accrual,
		0,
		time.Now(),
	).Scan(&id)
}

func (b *Balance) CheckBalance(userId int, withdrawRequest *model.BalanceWithdraw) error {
	var balance *float64
	err := b.db.Pool.QueryRow("select (sum(accrue) - sum(withdraw)) as balance  from balance where user_id = $1", userId).Scan(
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

func (b *Balance) WithDraw(userId int, withdrawRequest *model.BalanceWithdraw) error {
	var id int
	return b.db.Pool.QueryRow(
		"INSERT INTO balance (user_id, order_id, accrue, withdraw, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		userId,
		withdrawRequest.Order,
		0,
		withdrawRequest.Sum,
		time.Now(),
	).Scan(&id)
}

func (b *Balance) WithDrawals(userId int) ([]model.BalanceWithdrawalsResponse, error) {

	var withdraw model.BalanceWithdrawalsResponse
	var withdrawals []model.BalanceWithdrawalsResponse

	query := "SELECT order_id, withdraw, created_at FROM balance " +
		"WHERE withdraw > 0 and user_id = $1 ORDER BY created_at"

	rows, err := b.db.Pool.Query(query, userId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err = rows.Scan(&withdraw.Order, &withdraw.Sum, &withdraw.ProcessedAt); err != nil {
			return nil, err
		}
		//sum := fmt.Sprintf("%.2f", withdraw.Sum)
		//withdraw.Sum, _ = strconv.ParseFloat(sum, 64)
		//withdraw.Sum *= -1
		withdrawals = append(withdrawals, withdraw)
	}

	return withdrawals, nil
}
