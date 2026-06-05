package model

import "time"

// User is a registered account. PasswordHash is never serialized to clients.
type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // bcrypt hash; json:"-" keeps it out of every API response
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Role         string    `json:"role"` // enum: "customer" or "admin"; gates admin-only routes
	CreatedAt    time.Time `json:"created_at"`
}

// Address is a saved delivery address belonging to a user.
type Address struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	City      string `json:"city"`
	Street    string `json:"street"`
	House     string `json:"house"`
	Apartment string `json:"apartment"`
	ZipCode   string `json:"zip_code"`
	IsDefault bool   `json:"is_default"` // the address pre-selected at checkout
}
