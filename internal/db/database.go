package database

import (
	"context"
	"database/sql"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"github.com/vasiliyantufev/gophermart/internal/config"
)

type DB struct {
	Pool *sql.DB
}

func New(cfg *config.Config) (*DB, error) {
	pool, err := sql.Open("postgres", cfg.DatabaseURI)
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

func (db DB) CreateTablesMigration(cfg *config.Config) {

	driver, err := postgres.WithInstance(db.Pool, &postgres.Config{})
	if err != nil {
		log.Error(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		cfg.RootPath,

		"postgres", driver)
	if err != nil {
		log.Error(err)
	}
	if err = m.Up(); err != nil {
		log.Error(err)
	}
}
