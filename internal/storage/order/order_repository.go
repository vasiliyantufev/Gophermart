package order

import (
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
)

type OrderRepository interface {
	Create(*model.Order, database.DB) error
	FindByID(int, database.DB) (*model.Order, error)
}

type Order struct {
	orderRepository *OrderRepository
	db              *database.DB
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
		order.CreatedAt,
	).Scan(&order.ID)
}

func (o *Order) FindByLoginAndID(id int, user *model.User) (*model.Order, error) {

	order := &model.Order{}

	if err := o.db.Pool.QueryRow("SELECT * FROM orders where id=$1 and login=$2", id, user.Login).Scan(
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

func (o *Order) FindByID(id int) (*model.Order, error) {

	order := &model.Order{}

	if err := o.db.Pool.QueryRow("SELECT * FROM orders where id=$1", id).Scan(
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

func (o *Order) GetOrders(userId int) ([]model.Order, error) {

	var orders []model.Order
	var order model.Order

	//fmt.Print(userId)

	query := "SELECT * FROM orders"

	rows, err := o.db.Pool.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		if err = rows.Scan(&order.ID, &order.UserID, &order.OrderID, &order.CurrentStatus,
			&order.CreatedAt, &order.UpdatedAt,
		); err != nil {
			return orders, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}
