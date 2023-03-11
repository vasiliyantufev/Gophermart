package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/vasiliyantufev/gophermart/internal/api"
	"github.com/vasiliyantufev/gophermart/internal/config"
	database "github.com/vasiliyantufev/gophermart/internal/db"
)

func main() {

	cfg := config.New()

	store := sessions.NewCookieStore([]byte(cfg.SessionKey))

	db, _ := database.New(cfg)
	defer db.Close()

	log := logrus.New()
	log.SetLevel(cfg.DebugLevel)

	server := api.NewServer(log, cfg, db, store)

	r := chi.NewRouter()
	r.Mount("/", server.Route())

	server.StartServer(r, cfg, log)

}
