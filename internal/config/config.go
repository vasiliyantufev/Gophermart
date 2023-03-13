package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	Address              string    `env:"RUN_ADDRESS"`
	DatabaseUri          string    `env:"DATABASE_URI" envDefault:"host=localhost port=5432 user=postgres password=myPassword dbname=gophermart sslmode=disable"`
	AccrualSystemAddress string    `env:"ACCRUAL_SYSTEM_ADDRESS"`
	LogLevel             log.Level `env:"DEBUG_LEVEL" envDefault:"debug"`
	SessionKey           string    `env:"SESSION_KEY" envDefault:"secret"`
	TokenKey             string    `env:"TOKEN_KEY"   envDefault:"Q4RZDVti48qAsDw8u3NFLRScGyTMpbZ8tbA7Ubs8YJTZHMNBvw6vtCVrXbSHt5V1O-zf8OR35tbkApuri-TrHA"`
}

func New() *Config {

	cfg := Config{}

	// Установка флагов
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "Адрес и порт запуска сервиса")
	flag.StringVar(&cfg.DatabaseUri, "d", "", "Адрес подключения к базе данных")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", "", "Адрес системы расчёта начислений")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Info(cfg)

	return &cfg
}
