package service

import (
	"context"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
	"github.com/vasiliyantufev/gophermart/internal/config"
	"log"
)

type Accruer interface {
	StartWorkers(ctx context.Context)
	putOrdersWorker(ctx context.Context, urlPath string, client *resty.Client)
}

type accrual struct {
	cfg *config.Config
	log *logrus.Logger
}

func NewAccrual(log *logrus.Logger, cfg *config.Config) *accrual {
	return &accrual{log: log, cfg: cfg}
}

func (a accrual) StartWorkers(ctx context.Context, accruer Accruer) {

	client := resty.New()
	urlPath := "http://" + a.cfg.AccrualSystemAddress

	for i := 0; i < a.cfg.RateLimit; i++ {
		go accruer.putOrdersWorker(ctx, urlPath, client)
	}
}

func (a accrual) putOrdersWorker(ctx context.Context, urlPath string, client *resty.Client) {

	_, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(8888).
		Post(urlPath)
	if err != nil {
		log.Println(err)
		return
	}
}
