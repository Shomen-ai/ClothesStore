package model

import "time"

type PromoCode struct {
	ID               int64      `json:"id"`
	Code             string     `json:"code"`
	DiscountType     string     `json:"discount_type"`
	DiscountValue    float64    `json:"discount_value"`
	MaxActivations   *int       `json:"max_activations"`
	ActivationsCount int        `json:"activations_count"`
	ExpiresAt        *time.Time `json:"expires_at"`
	IsActive         bool       `json:"is_active"`
	CreatedAt        time.Time  `json:"created_at"`
}
