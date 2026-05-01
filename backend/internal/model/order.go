package model

import "time"

type OrderItem struct {
	ID            int64   `json:"id"`
	OrderID       int64   `json:"order_id"`
	ProductID     int64   `json:"product_id"`
	ProductSizeID int64   `json:"product_size_id"`
	Quantity      int     `json:"quantity"`
	PriceAtOrder  float64 `json:"price_at_order"`
}

type Order struct {
	ID             int64       `json:"id"`
	UserID         int64       `json:"user_id"`
	AddressID      int64       `json:"address_id"`
	PromoCodeID    *int64      `json:"promo_code_id"`
	Status         string      `json:"status"`
	TotalPrice     float64     `json:"total_price"`
	DiscountAmount float64     `json:"discount_amount"`
	CreatedAt      time.Time   `json:"created_at"`
	Items          []OrderItem `json:"items,omitempty"`
}
