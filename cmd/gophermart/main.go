package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"github.com/vasiliyantufev/gophermart/internal/api"
	"github.com/vasiliyantufev/gophermart/internal/config"
	database "github.com/vasiliyantufev/gophermart/internal/db"
	"github.com/vasiliyantufev/gophermart/internal/service"
)

func main() {

	cfg := config.New()

	store := sessions.NewCookieStore([]byte(cfg.SessionKey))

	keyb := service.DecodeKey(cfg.TokenKey)
	jwt := service.NewJwt(keyb)

	log := logrus.New()
	log.SetLevel(cfg.LogLevel)

	db, err := database.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	server := api.NewServer(log, cfg, db, store, jwt)

	r := chi.NewRouter()
	r.Mount("/", server.Route())

	server.StartServer(r, cfg, log)
}
