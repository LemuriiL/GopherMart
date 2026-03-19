package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress string
	JWTSecret  string
}

func New() *Config {
	cfg := &Config{
		RunAddress: ":8080",
		JWTSecret:  "supersecretkey",
	}

	flag.StringVar(&cfg.RunAddress, "a", getEnv("RUN_ADDRESS", cfg.RunAddress), "server address")
	flag.StringVar(&cfg.JWTSecret, "s", getEnv("JWT_SECRET", cfg.JWTSecret), "jwt secret")
	flag.Parse()

	return cfg
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
