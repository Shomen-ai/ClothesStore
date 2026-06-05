package service

import (
	"errors"
	"time"
	"clothes-store/internal/model"
)

// ValidatePromo enforces the business rules that decide whether a promo code
// may be used right now: it must be active, not past its expiry (when one is
// set), and below its activation cap (when one is set). nil means usable.
func ValidatePromo(p *model.PromoCode) error {
	if !p.IsActive {
		return errors.New("promo code is inactive")
	}
	// ExpiresAt is optional (nil = no expiry); reject only once it's in the past.
	if p.ExpiresAt != nil && p.ExpiresAt.Before(time.Now()) {
		return errors.New("promo code has expired")
	}
	// MaxActivations is optional (nil = unlimited); reject once the cap is hit.
	if p.MaxActivations != nil && p.ActivationsCount >= *p.MaxActivations {
		return errors.New("promo code activation limit reached")
	}
	return nil
}

// ApplyDiscount computes the discount amount (not the final total) for the
// given order subtotal. "percent" takes a share of the total; "fixed" subtracts
// a flat value but is clamped to the total so the order can never go negative.
// An unknown discount type yields no discount.
func ApplyDiscount(p *model.PromoCode, total float64) float64 {
	switch p.DiscountType {
	case "percent":
		return total * p.DiscountValue / 100
	case "fixed":
		// A fixed discount larger than the subtotal is capped at the subtotal.
		if p.DiscountValue > total {
			return total
		}
		return p.DiscountValue
	}
	return 0
}
