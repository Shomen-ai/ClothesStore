// Package db wires up the PostgreSQL connection pool used by the repositories.
package db

import (
	"database/sql"
	"fmt"
	"time"
	_ "github.com/lib/pq" // register the "postgres" driver via blank import
)

// Connect opens a PostgreSQL connection pool from connStr, verifies it with a
// Ping (sql.Open is lazy and does not connect on its own), tunes the pool
// limits, and returns the ready-to-use *sql.DB.
func Connect(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	if err := db.Ping(); err != nil { // force a real connection to surface bad DSN/credentials early
		return nil, fmt.Errorf("ping: %w", err)
	}
	db.SetMaxOpenConns(25)                  // cap concurrent open connections
	db.SetMaxIdleConns(5)                   // keep a few warm idle connections
	db.SetConnMaxLifetime(30 * time.Minute) // recycle connections after 30m
	db.SetConnMaxIdleTime(5 * time.Minute)  // drop connections idle longer than 5m
	return db, nil
}
