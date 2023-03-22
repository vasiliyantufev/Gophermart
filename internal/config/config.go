package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Address     string `env:"RUN_ADDRESS"`
	DatabaseURI string `env:"DATABASE_URI"`
	//DatabaseUri          string    `env:"DATABASE_URI" envDefault:"postgresql://postgres:postgres@postgres:5432/postgres?sslmode=disable"`
	AccrualSystemAddress string    `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel             log.Level `env:"DEBUG_LEVEL" envDefault:"debug"`
	RootPath             string    `env:"ROOT_PATH" envDefault:"file://./migrations"`
	//	RateLimit            int       `env:"RATE_LIMIT"`
}

func New() *Config {

	cfg := Config{}

	//flag.StringVar(&cfg.Address, "a", "localhost:8088", "Адрес и порт запуска сервиса")
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Адрес и порт запуска сервиса")
	//flag.StringVar(&cfg.DatabaseURI, "d", "host=localhost port=5432 user=postgres password=myPassword dbname=gophermart sslmode=disable", "Адрес подключения к базе данных")
	flag.StringVar(&cfg.DatabaseURI, "d", "postgresql://postgres:postgres@postgres:5432/postgres?sslmode=disable", "Адрес подключения к базе данных")
	//flag.StringVar(&cfg.AccrualSystemAddress, "r", "localhost:8080", "Адрес и порт запуска системы расчёта начислений")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Адрес и порт запуска системы расчёта начислений")

	//	flag.IntVar(&cfg.RateLimit, "l", 2, "Количество одновременно исходящих запросов на сервер")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Info(cfg)

	return &cfg
}
