package model

import "time"

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type Address struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	City      string `json:"city"`
	Street    string `json:"street"`
	House     string `json:"house"`
	Apartment string `json:"apartment"`
	ZipCode   string `json:"zip_code"`
	IsDefault bool   `json:"is_default"`
}
