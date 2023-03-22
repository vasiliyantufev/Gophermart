package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	_ "github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/vasiliyantufev/gophermart/internal/api/accrual"
	"github.com/vasiliyantufev/gophermart/internal/api/server"
	"github.com/vasiliyantufev/gophermart/internal/config"
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/storage/repositories/balance"
	"github.com/vasiliyantufev/gophermart/internal/storage/repositories/order"
	"github.com/vasiliyantufev/gophermart/internal/storage/repositories/token"
	"github.com/vasiliyantufev/gophermart/internal/storage/repositories/user"
	"os/signal"
	"syscall"
)

func main() {

	cfg := config.New()

	log := logrus.New()
	log.SetLevel(cfg.LogLevel)

	db, err := database.New(cfg)
	if err != nil {
		log.Fatal(err)
	} else {
		defer db.Close()
		db.CreateTablesMigration(cfg)
	}

	userRepository := user.New(db)
	orderRepository := order.New(db)
	balanceRepository := balance.New(db)
	tokenRepository := token.New(db)

	accrual := accrual.NewAccrual(log, cfg, db, orderRepository, balanceRepository)
	server := server.NewServer(log, cfg, db, userRepository, orderRepository, balanceRepository, tokenRepository)

	r := chi.NewRouter()
	r.Mount("/", server.Route())

	ctx, cnl := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cnl()

	go server.StartServer(r, cfg, log)
	go accrual.StartWorkers(ctx, accrual)

	<-ctx.Done()
	log.Error("gophermart shutdown on signal with:", ctx.Err())
}
