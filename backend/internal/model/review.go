package model

import "time"

// Review is a customer's rating and text feedback for a product.
type Review struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	ProductID int64     `json:"product_id"`
	Rating    int       `json:"rating"`              // star rating, 1-5
	Text      string    `json:"text"`                // free-form review body
	IsHidden  bool      `json:"is_hidden,omitempty"` // moderation flag; omitempty keeps it out of public payloads when false
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Optional joined fields, populated by list queries.
	UserName    string `json:"user_name,omitempty"`    // joined from users; omitted when not loaded
	UserEmail   string `json:"user_email,omitempty"`   // joined from users; omitted when not loaded
	ProductName string `json:"product_name,omitempty"` // joined from products; omitted when not loaded
}

// ReviewSummary is the aggregate shown next to the reviews list on a product page.
type ReviewSummary struct {
	AvgRating float64 `json:"avg_rating"` // mean of all visible ratings (float64)
	Count     int     `json:"count"`      // number of visible reviews
}
