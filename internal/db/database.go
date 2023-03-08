package database

import (
	"context"
	"database/sql"
	log "github.com/sirupsen/logrus"
	"github.com/vasiliyantufev/gophermart/internal/config"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	Pool *sql.DB
}

func New(cfg *config.Config) (*DB, error) {
	pool, err := sql.Open("postgres", cfg.DatabaseUri)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cnl := context.WithTimeout(context.Background(), 10*time.Second)
	defer cnl()

	if err := pool.PingContext(ctx); err != nil {
		return nil, err

	}
	return &DB{Pool: pool}, nil
}

func (db DB) Close() error {
	return db.Pool.Close()
}

func (db DB) Ping() error {
	if err := db.Pool.Ping(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}
