package main

import (
	"context"
	"github.com/go-chi/chi/v5"
	_ "github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/vasiliyantufev/gophermart/internal/api/server"
	"github.com/vasiliyantufev/gophermart/internal/config"
	database "github.com/vasiliyantufev/gophermart/internal/db"
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

	server := server.NewServer(log, cfg, db)
	//accrual := accrual.NewAccrual(log, cfg)

	r := chi.NewRouter()
	r.Mount("/", server.Route())

	ctx, cnl := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer cnl()

	go server.StartServer(r, cfg, log)
	//go accrual.StartWorkers(ctx, accrual)

	<-ctx.Done()
	log.Error("gophermart shutdown on signal with:", ctx.Err())
}
