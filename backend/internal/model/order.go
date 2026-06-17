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
	// Display-only fields hydrated by the detail query (joins) so the order card can
	// show the product without extra requests; empty on list responses.
	ProductName string `json:"product_name,omitempty"`
	TypeName    string `json:"type_name,omitempty"`
	Size        string `json:"size,omitempty"`
	ImagePath   string `json:"image_path,omitempty"`
}

// Order is a customer's placed order with its computed money totals and lines.
type Order struct {
	ID             int64       `json:"id"`
	UserID         int64       `json:"user_id"`
	AddressID      int64       `json:"address_id"`
	PromoCodeID    *int64      `json:"promo_code_id"`   // nullable FK: nil when no promo code was applied
	Status         string      `json:"status"`          // enum: pending, paid, shipped, delivered, cancelled
	TotalPrice     float64     `json:"total_price"`     // amount due in rubles: subtotal − discount + delivery
	DiscountAmount float64     `json:"discount_amount"` // discount applied in rubles
	DeliveryMethod string      `json:"delivery_method"` // courier | post | pickup
	DeliveryCost   float64     `json:"delivery_cost"`   // shipping fee in rubles (server-computed)
	RecipientName  string      `json:"recipient_name"`  // who receives the parcel
	PaymentMethod  string      `json:"payment_method"`  // card_online | on_delivery
	PaymentStatus  string      `json:"payment_status"`  // unpaid | paid | on_delivery
	CreatedAt      time.Time   `json:"created_at"`
	Items          []OrderItem `json:"items,omitempty"`   // hydrated by detail queries only; omitted from list responses
	Address        *Address    `json:"address,omitempty"` // delivery address, hydrated by detail query
}
