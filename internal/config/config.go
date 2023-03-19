package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	//Address              string    `env:"RUN_ADDRESS"`
	Address              string    `env:"RUN_ADDRESS"`
	DatabaseUri          string    `env:"DATABASE_URI" envDefault:"host=localhost port=5432 user=postgres password=myPassword dbname=gophermart sslmode=disable"`
	AccrualSystemAddress string    `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel             log.Level `env:"DEBUG_LEVEL" envDefault:"debug"`
	RootPath             string    `env:"ROOT_PATH" envDefault:"file://../../"`
	//	RateLimit            int       `env:"RATE_LIMIT"`
}

func New() *Config {

	cfg := Config{}

	flag.StringVar(&cfg.Address, "a", "127.0.0.1:8088", "Адрес и порт запуска сервиса")
	flag.StringVar(&cfg.DatabaseUri, "d", "", "Адрес подключения к базе данных")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Адрес системы расчёта начислений")
	//	flag.IntVar(&cfg.RateLimit, "l", 2, "Количество одновременно исходящих запросов на сервер")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Info(cfg)

	return &cfg
}
