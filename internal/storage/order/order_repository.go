package order

import (
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
	Servicer Servicer
	db       *database.DB
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

	query := "SELECT * FROM orders where user_id = $1"

	rows, err := o.db.Pool.Query(query, userId)
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

func (o *Order) CheckOrder(orderID *model.OrderID) error {

	//orderID := &model.OrderID{}
	//if err := json.NewDecoder(r.Body).Decode(orderID); err != nil {
	//	s.log.Error(err)
	//	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//	return
	//}
	//
	//o, _ := s.orderRepository.Servicer.FindByID(orderID.Order)
	//if o == nil {
	//	s.log.Error("Order is not registered in the billing system")
	//	http.Error(w, "Order is not registered in the billing system", http.StatusNoContent)
	//	return
	//}
	//
	//if orderID.Status == "INVALID" {
	//	err := s.orderRepository.Servicer.Update(orderID)
	//	if err != nil {
	//		s.log.Error(err)
	//		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//	}
	//}
	//if orderID.Status == "PROCESSED" {
	//	user := r.Context().Value("userPayloadCtx").(*model.User)
	//	err := s.orderRepository.Servicer.Update(orderID)
	//	if err != nil {
	//		s.log.Error(err)
	//		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//	}
	//	err = s.balanceRepository.Balancer.Accrue(user.ID, orderID)
	//	if err != nil {
	//		s.log.Error(err)
	//		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	//	}
	//}
	//
	//s.log.Info("Successful request processing")
	//http.Error(w, "Successful request processing", http.StatusOK)
	return nil
}
