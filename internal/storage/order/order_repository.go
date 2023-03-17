package order

import (
	"github.com/sirupsen/logrus"
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
	"time"
)

type Servicer interface {
	Create(order *model.Order) error
	Update(orderID *model.OrderID) error
	FindByLoginAndID(id int, user *model.User) (*model.Order, error)
	FindByID(id int) (*model.Order, error)
	GetOrders(userId int) ([]model.Order, error)
	CheckOrder(orderID *model.OrderID) error
}

type Order struct {
	db *database.DB
}

func New(db *database.DB) *Order {
	return &Order{
		db: db,
	}
}

func (o *Order) Create(order *model.Order) error {

	return o.db.Pool.QueryRow(
		"INSERT INTO orders (user_id, order_id, current_status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		order.UserID,
		order.OrderID,
		order.CurrentStatus,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&order.ID)
}

func (o *Order) Update(orderID *model.OrderID) error {

	var id int
	return o.db.Pool.QueryRow("UPDATE orders SET current_status = $2, updated_at = $3 WHERE id = $1 RETURNING id;",
		orderID.Order, orderID.Status, time.Now()).Scan(&id)
}

func (o *Order) FindByOrderIDAndUserID(orderId int, userId int) (*model.Order, error) {

	order := &model.Order{}

	if err := o.db.Pool.QueryRow("SELECT * FROM orders where order_id=$1 and user_id=$2", orderId, userId).Scan(
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

func (o *Order) FindByOrderID(orderId int) (*model.Order, error) {

	order := &model.Order{}

	if err := o.db.Pool.QueryRow("SELECT * FROM orders where order_id=$1", orderId).Scan(
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

//func (o *Order) GetOrders(userId int) ([]model.Order, error) {
//
//	var orders []model.Order
//	var order model.Order
//
//	query := "SELECT * FROM orders where user_id = $1"
//
//	rows, err := o.db.Pool.Query(query, userId)
//	if err != nil {
//		return nil, err
//	}
//
//	for rows.Next() {
//		if err = rows.Scan(&order.ID, &order.UserID, &order.OrderID, &order.CurrentStatus,
//			&order.CreatedAt, &order.UpdatedAt,
//		); err != nil {
//			return nil, err
//		}
//		orders = append(orders, order)
//	}
//
//	return orders, nil
//}

func (o *Order) GetOrders(userId int) ([]model.OrderResponse, error) {

	logrus.Info(userId)

	var orders []model.OrderResponse
	var order model.OrderResponse

	query := "SELECT orders.order_id as number, " +
		"orders.current_status as status, " +
		//"COALESCE(sum(balance.delta), 0) as accrual, " +
		"sum(balance.delta) as accrual, " +
		"orders.updated_at as uploaded_at " +
		"from orders " +
		"LEFT JOIN balance ON balance.order_id = orders.order_id " +
		"where orders.user_id = $1 " +
		//"GROUP BY number, uploaded_at"
		"GROUP BY number, status, uploaded_at"

	logrus.Info(query)

	rows, err := o.db.Pool.Query(query, userId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err = rows.Scan(&order.Number, &order.Status /**/, &order.Accrual, &order.UploadedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (o *Order) GetOrdersToAccrual() ([]model.Order, error) {

	var orders []model.Order
	var order model.Order

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

	return orders, nil
}
