package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/vasiliyantufev/gophermart/internal/config"
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/model"
	"github.com/vasiliyantufev/gophermart/internal/storage/balance"
	"github.com/vasiliyantufev/gophermart/internal/storage/order"
	"io"
	"log"
	"net/http"
	"strconv"
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
}

func NewAccrual(log *logrus.Logger, cfg *config.Config) *accrual {
	return &accrual{log: log, cfg: cfg}
}

func (a accrual) StartWorkers(ctx context.Context, accruar *accrual) {

	a.orderRepository = order.New(a.db)
	urlPath := "http://" + a.cfg.AccrualSystemAddress
	accruar.putOrdersWorker(ctx, urlPath)
}

func (a accrual) putOrdersWorker(ctx context.Context, urlPath string) {

	ticker := time.NewTicker(0)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("accrual stopped by ctx")
			return
		case <-ticker.C:
			orders, err := a.orderRepository.GetOrdersToAccrual()
			if err != nil {
				log.Println(err)
				return
			}

			for _, order := range orders {
				a.log.Info("Get order: " + strconv.Itoa(order.OrderID))
				urlPath = "http://" + a.cfg.AccrualSystemAddress
				go a.makeGetRequest(order.ID, urlPath)
			}
		}
	}

}

func (a accrual) makeGetRequest(id int, url string) {

	var body []byte

	ctx := context.Background()

	var orderID *model.OrderID

	urlOrder := url + "/" + strconv.Itoa(id)

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

	a.CheckOrder(orderID, ctx)

}

func (a accrual) CheckOrder(orderID *model.OrderID, ctx context.Context) error {

	o, _ := a.orderRepository.FindByOrderID(orderID.Order)
	if o == nil {
		return errors.New("Order is not registered in the billing system")
	}

	if orderID.Status == "INVALID" {
		err := a.orderRepository.Update(orderID)
		if err != nil {
			a.log.Error(err)
		}
	}
	if orderID.Status == "PROCESSED" {
		user := ctx.Value("userPayloadCtx").(*model.User)
		err := a.orderRepository.Update(orderID)
		if err != nil {
			a.log.Error(err)
		}
		err = a.balanceRepository.Accrue(user.ID, orderID)
		if err != nil {
			a.log.Error(err)
		}
	}
	return nil
}
