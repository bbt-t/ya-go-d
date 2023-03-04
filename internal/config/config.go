package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	ServerAddress  string `env:"RUN_ADDRESS"`
	DSN            string `env:"DATABASE_URI"`
	AccrualAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	SecretKey      string `env:"SECRET_KEY" envDefault:"OlOlO"`
	WaitingTime    time.Duration
	Workers        int
}

func NewConfig() *Config {
	/*
		Initialize a new config.
		(env + flag)
	*/
	cfg := Config{
		WaitingTime: 500 * time.Millisecond,
		Workers:     2,
	}

	flag.StringVar(&cfg.ServerAddress, "a", ":8080", "Server address")
	flag.StringVar(&cfg.DSN, "d", "", "Postgres DSN (url)")
	flag.StringVar(&cfg.AccrualAddress, "r", "http://127.0.0.1:8088", "Accrual address")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}

	return &cfg
}
