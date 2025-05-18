package config

import (
	"log"
	"os"
)

type Config struct {
	DBUrl string // e.g postgres://admin:password@localhost:5432/postgres
}

func Load() Config {
	cfg := Config{
		DBUrl: os.Getenv("DB_URL"),
	}

	if cfg.DBUrl == "" {
		log.Fatal("DB_URL is not set")
	}

	return cfg
}
