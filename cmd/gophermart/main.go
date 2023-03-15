package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	_ "github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/vasiliyantufev/gophermart/internal/api"
	"github.com/vasiliyantufev/gophermart/internal/config"
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/service"
	"os/signal"
	"syscall"
)

func main() {

	cfg := config.New()

	//store := sessions.NewCookieStore([]byte(cfg.SessionKey))
	//keyb := service.DecodeKey(cfg.TokenKey)

	key := service.GenerateKey()
	keyb := service.DecodeKey(key)

	jwt := service.NewJwt(keyb)

	log := logrus.New()
	log.SetLevel(cfg.LogLevel)

	db, err := database.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	server := api.NewServer(log, cfg, db /*store,*/, jwt)
	accrual := api.NewAccrual(log, cfg)

	r := chi.NewRouter()
	r.Mount("/", server.Route())

	ctx, cnl := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cnl()

	server.StartServer(r, cfg, log)
	accrual.StartWorkers(ctx, accrual)

	<-ctx.Done()
	log.Println("gophermart shutdown on signal with:", ctx.Err())
}
