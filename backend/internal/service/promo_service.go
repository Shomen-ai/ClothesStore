package service

import (
	"errors"
	"time"
	"clothes-store/internal/model"
)

func ValidatePromo(p *model.PromoCode) error {
	if !p.IsActive {
		return errors.New("promo code is inactive")
	}
	if p.ExpiresAt != nil && p.ExpiresAt.Before(time.Now()) {
		return errors.New("promo code has expired")
	}
	if p.MaxActivations != nil && p.ActivationsCount >= *p.MaxActivations {
		return errors.New("promo code activation limit reached")
	}
	return nil
}

func ApplyDiscount(p *model.PromoCode, total float64) float64 {
	switch p.DiscountType {
	case "percent":
		return total * p.DiscountValue / 100
	case "fixed":
		if p.DiscountValue > total {
			return total
		}
		return p.DiscountValue
	}
	return 0
}
