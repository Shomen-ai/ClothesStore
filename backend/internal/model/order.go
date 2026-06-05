package model

import "time"

// OrderItem is one line of an order: a specific product/size and how many were
// bought, with the price frozen at purchase time.
type OrderItem struct {
	ID            int64   `json:"id"`
	OrderID       int64   `json:"order_id"`
	ProductID     int64   `json:"product_id"`
	ProductSizeID int64   `json:"product_size_id"`
	Quantity      int     `json:"quantity"`
	PriceAtOrder  float64 `json:"price_at_order"` // per-unit price captured at order time (rubles); insulates the line total from later price changes
}

// Order is a customer's placed order with its computed money totals and lines.
type Order struct {
	ID             int64       `json:"id"`
	UserID         int64       `json:"user_id"`
	AddressID      int64       `json:"address_id"`
	PromoCodeID    *int64      `json:"promo_code_id"`   // nullable FK: nil when no promo code was applied
	Status         string      `json:"status"`          // enum: pending, paid, shipped, delivered, cancelled
	TotalPrice     float64     `json:"total_price"`     // amount due in rubles after the discount is subtracted
	DiscountAmount float64     `json:"discount_amount"` // discount applied in rubles
	CreatedAt      time.Time   `json:"created_at"`
	Items          []OrderItem `json:"items,omitempty"` // hydrated by detail queries only; omitted from list responses
}
