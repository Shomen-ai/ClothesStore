package model

import "time"

// PromoCode is a discount code that can be applied to an order at checkout.
type PromoCode struct {
	ID               int64      `json:"id"`
	Code             string     `json:"code"`            // the code customers type at checkout
	DiscountType     string     `json:"discount_type"`   // enum: "percent" or "fixed"
	DiscountValue    float64    `json:"discount_value"`  // percent (0-100) or fixed rubles depending on DiscountType
	MaxActivations   *int       `json:"max_activations"` // nullable: nil means unlimited uses
	ActivationsCount int        `json:"activations_count"`
	ExpiresAt        *time.Time `json:"expires_at"` // nullable: nil means the code never expires
	IsActive         bool       `json:"is_active"`  // false disables the code regardless of other limits
	CreatedAt        time.Time  `json:"created_at"`
}
