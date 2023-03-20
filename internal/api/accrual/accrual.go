package accrual

import (
	"context"
	"encoding/json"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/vasiliyantufev/gophermart/internal/config"
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
	"github.com/vasiliyantufev/gophermart/internal/storage/errors"
	"github.com/vasiliyantufev/gophermart/internal/storage/repositories/balance"
	"github.com/vasiliyantufev/gophermart/internal/storage/repositories/order"
	"github.com/vasiliyantufev/gophermart/internal/storage/statuses"
	"io"
	"net/http"
	"time"
)

type Accruer interface {
	StartWorkers(ctx context.Context)
	putOrdersWorker(ctx context.Context, urlPath string)
	makeGetRequest(client *resty.Client, id int, url string)
}

type accrual struct {
	cfg               *config.Config
	db                *database.DB
	log               *logrus.Logger
	orderRepository   *order.Order
	balanceRepository *balance.Balance
	urlPath           string
}

func NewAccrual(log *logrus.Logger, cfg *config.Config, db *database.DB, orderRepository *order.Order, balanceRepository *balance.Balance) *accrual {
	return &accrual{log: log, cfg: cfg, db: db, orderRepository: orderRepository, balanceRepository: balanceRepository}
}

func (a accrual) StartWorkers(ctx context.Context, accruar *accrual) {

	a.urlPath = "http://" + a.cfg.AccrualSystemAddress
	accruar.putOrdersWorker(ctx)
}

func (a accrual) putOrdersWorker(ctx context.Context) {

	// TODO: change ticker time
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			a.log.Error("accrual stopped by ctx")
			return
		case <-ticker.C:
			orders, err := a.orderRepository.GetOrdersToAccrual()
			if err != nil {
				a.log.Error(err)
				return
			}

			for _, order := range orders {
				go a.makeGetRequest(order.OrderID)
			}
		}
	}
}

func (a accrual) makeGetRequest(id string) {

	var body []byte
	var orderID *model.OrderResponseAccrual

	urlOrder := a.urlPath + "/api/orders/" + id
	r, err := http.Get(urlOrder)
	if err != nil {
		return
	}

	body, err = io.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, orderID)
	if err != nil {
		return
	}

	a.CheckOrder(orderID)

}

func (a accrual) CheckOrder(orderID *model.OrderResponseAccrual) error {

	o, _ := a.orderRepository.FindByOrderID(orderID.Order)
	if o == nil {
		return errors.ErrNotRegistered
	}

	if orderID.Status == statuses.Invalid {
		userID, err := a.orderRepository.Update(orderID)
		if err != nil {
			a.log.Error(err)
		}
		a.log.Info("Get order: " + string(userID))
	}
	if orderID.Status == statuses.Processed {
		userID, err := a.orderRepository.Update(orderID)
		if err != nil {
			a.log.Error(err)
		}
		a.log.Info("Get order: " + string(userID))
		err = a.balanceRepository.Accrue(userID, orderID)
		if err != nil {
			a.log.Error(err)
		}
	}
	return nil
}
