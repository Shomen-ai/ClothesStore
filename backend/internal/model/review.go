package model

import "time"

type Review struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	Rating    int       `json:"rating"`
	Text      string    `json:"text"`
	IsHidden  bool      `json:"is_hidden,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Optional joined fields, populated by list queries.
	UserName    string `json:"user_name,omitempty"`
	UserEmail   string `json:"user_email,omitempty"`
	ProductName string `json:"product_name,omitempty"`
}

// ReviewSummary is the aggregate shown next to the reviews list on a product page.
type ReviewSummary struct {
	AvgRating float64 `json:"avg_rating"`
	Count     int     `json:"count"`
}
