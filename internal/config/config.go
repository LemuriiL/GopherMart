package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress     string
	JWTSecret      string
	AccrualAddress string
	DatabaseURI    string
}

func New() *Config {
	cfg := &Config{
		RunAddress:     ":8080",
		JWTSecret:      "supersecretkey",
		AccrualAddress: "http://localhost:8081",
		DatabaseURI:    "postgres://postgres:postgres@localhost:5432/gophermart?sslmode=disable",
	}

	flag.StringVar(&cfg.RunAddress, "a", getEnv("RUN_ADDRESS", cfg.RunAddress), "server address")
	flag.StringVar(&cfg.JWTSecret, "s", getEnv("JWT_SECRET", cfg.JWTSecret), "jwt secret")
	flag.StringVar(&cfg.AccrualAddress, "r", getEnv("ACCRUAL_SYSTEM_ADDRESS", cfg.AccrualAddress), "accrual address")
	flag.StringVar(&cfg.DatabaseURI, "d", getEnv("DATABASE_URI", cfg.DatabaseURI), "database uri")

	flag.Parse()

	return cfg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
