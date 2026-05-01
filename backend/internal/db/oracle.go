package db

import (
	"database/sql"
	"fmt"
	_ "github.com/godror/godror"
)

func Connect(connStr string) (*sql.DB, error) {
	db, err := sql.Open("godror", connStr)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return db, nil
}
