package order

import (
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
	"time"
)

type Servicer interface {
	Create(order *model.OrderDB) error
	Update(orderID *model.OrderResponseAccrual) error
	FindByOrderIDAndUserID(orderID string, userID int) (*model.OrderDB, error)
	FindByOrderID(orderID string) (*model.OrderDB, error)
	GetOrders(userID int) ([]model.OrdersResponseGophermart, error)
	GetOrdersToAccrual() ([]model.OrderDB, error)
}

type Order struct {
	db *database.DB
}

func New(db *database.DB) *Order {
	return &Order{
		db: db,
	}
}

func (o *Order) Create(order *model.OrderDB) error {

	return o.db.Pool.QueryRow(
		"INSERT INTO orders (user_id, order_id, current_status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		order.UserID,
		order.OrderID,
		order.CurrentStatus,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&order.ID)
}

func (o *Order) Update(orderID model.OrderResponseAccrual) (int, error) {

	var userID int
	if err := o.db.Pool.QueryRow("UPDATE orders SET current_status = $1, updated_at = $2 WHERE order_id = $3 RETURNING user_id;",
		orderID.Status, time.Now(), orderID.Order).Scan(&userID); err != nil {
		return userID, err
	}

	return userID, nil
}

func (o *Order) FindByOrderIDAndUserID(orderID string, userID int) (*model.OrderDB, error) {

	order := &model.OrderDB{}

	if err := o.db.Pool.QueryRow("SELECT * FROM orders where order_id=$1 and user_id=$2", orderID, userID).Scan(
		&order.ID,
		&order.UserID,
		&order.OrderID,
		&order.CurrentStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return order, nil
}

func (o *Order) FindByOrderID(orderID string) (*model.OrderDB, error) {

	order := &model.OrderDB{}

	if err := o.db.Pool.QueryRow("SELECT * FROM orders where order_id=$1", orderID).Scan(
		&order.ID,
		&order.UserID,
		&order.OrderID,
		&order.CurrentStatus,
		&order.CreatedAt,
		&order.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return order, nil
}

func (o *Order) GetOrders(userID int) ([]model.OrdersResponseGophermart, error) {

	var orders []model.OrdersResponseGophermart
	var order model.OrdersResponseGophermart

	query := "SELECT orders.order_id as number, " +
		"orders.current_status as status, " +
		"sum(balance.accrue) as accrual, " +
		"orders.updated_at as uploaded_at " +
		"from orders " +
		"LEFT JOIN balance ON balance.order_id = orders.order_id " +
		"where orders.user_id = $1 " +
		"GROUP BY number, status, uploaded_at"

	rows, err := o.db.Pool.Query(query, userID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return orders, nil
}

func (o *Order) GetOrdersToAccrual() ([]model.OrderDB, error) {

	var orders []model.OrderDB
	var order model.OrderDB

	query := "SELECT * FROM orders where current_status != 'INVALID' and current_status != 'PROCESSED'"

	rows, err := o.db.Pool.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err = rows.Scan(&order.ID, &order.UserID, &order.OrderID, &order.CurrentStatus,
			&order.CreatedAt, &order.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if rows.Err() != nil {
		return nil, err
	}

	return orders, nil
}
