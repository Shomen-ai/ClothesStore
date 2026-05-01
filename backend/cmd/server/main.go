package main

import (
	"log"
	"clothes-store/internal/config"
)

func main() {
	cfg := config.Load()
	if cfg.DBConnStr == "" {
		log.Fatal("DB_CONN_STR is required")
	}
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	log.Printf("Starting server on :%s", cfg.Port)
}
