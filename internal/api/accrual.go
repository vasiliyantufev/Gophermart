package api

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/vasiliyantufev/gophermart/internal/config"
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/storage/order"
	"log"
	"strconv"
	"time"
)

type Accruer interface {
	StartWorkers(ctx context.Context)
	putOrdersWorker(ctx context.Context, urlPath string)
	makeGetRequest(client *resty.Client, id int, url string)
}

type accrual struct {
	cfg             *config.Config
	db              *database.DB
	log             *logrus.Logger
	orderRepository *order.Order
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

	client := resty.New()

	ticker := time.NewTicker(10)
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
				go a.makeGetRequest(client, order.ID, urlPath)
			}
		}
	}

}

func (a accrual) makeGetRequest(client *resty.Client, id int, url string) {

	urlOrder := url + "/" + strconv.Itoa(id)

	//orderID := &model.OrderID{}

	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(id).
		Get(urlOrder)
	if err != nil {
		log.Println(err)
		return
	}

	//a.orderRepository.Servicer.CheckOrder(orderID)

	//return order
}
