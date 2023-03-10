package order

import (
	"fmt"
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
)

type OrderRepository interface {
	Create(*model.Order, database.DB) error
	FindByID(int, database.DB) (*model.Order, error)
}

type Order struct {
	orderRepository *OrderRepository
}

func (o *Order) Create(order *model.Order, db *database.DB) error {

	return db.Pool.QueryRow(
		"INSERT INTO orders (user_id, order_number, status, accrual, uploaded_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id",
		order.UserID,
		order.OrderNumber,
		order.Status,
		order.Accrual,
		order.CreatedAt,
		order.UploadedAt,
	).Scan(&order.ID)
}

func (o *Order) FindByID(id int, db *database.DB) (*model.Order, error) {

	order := &model.Order{}

	if err := db.Pool.QueryRow("SELECT * FROM orders where id=$1", id).Scan(
		&order.UserID,
		&order.OrderNumber,
		&order.Status,
		&order.Accrual,
		&order.CreatedAt,
		&order.UploadedAt,
	); err != nil {
		return nil, err
	}
	return order, nil
}

func (o *Order) GetOrders(userId int, db *database.DB) ([]model.Order, error) {

	var orders []model.Order
	var order model.Order

	fmt.Print(userId)

	query := "SELECT * FROM orders"

	rows, err := db.Pool.Query(query)

	if err != nil {
		return orders, nil
	}

	for rows.Next() {
		if err = rows.Scan(&order.ID, &order.UserID, &order.OrderNumber, &order.Status,
			&order.Accrual, &order.CreatedAt, &order.UploadedAt,
		); err != nil {
			return orders, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}
