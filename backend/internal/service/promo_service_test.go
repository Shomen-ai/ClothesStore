package service_test

import (
	"testing"
	"time"
	"clothes-store/internal/model"
	"clothes-store/internal/service"
)

func promoWith(t string, v float64) *model.PromoCode {
	return &model.PromoCode{Code: "TEST", DiscountType: t, DiscountValue: v, IsActive: true}
}

func TestApplyPromo_Percent(t *testing.T) {
	p := promoWith("percent", 20)
	discount := service.ApplyDiscount(p, 1000)
	if discount != 200 {
		t.Errorf("want 200, got %f", discount)
	}
}

func TestApplyPromo_Fixed(t *testing.T) {
	p := promoWith("fixed", 150)
	discount := service.ApplyDiscount(p, 1000)
	if discount != 150 {
		t.Errorf("want 150, got %f", discount)
	}
}

func TestApplyPromo_Fixed_CannotExceedTotal(t *testing.T) {
	p := promoWith("fixed", 2000)
	discount := service.ApplyDiscount(p, 500)
	if discount != 500 {
		t.Errorf("fixed discount must not exceed total, got %f", discount)
	}
}

func TestValidatePromo_Expired(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	p := &model.PromoCode{IsActive: true, ExpiresAt: &past}
	if err := service.ValidatePromo(p); err == nil {
		t.Error("expected error for expired promo")
	}
}

func TestValidatePromo_Inactive(t *testing.T) {
	p := &model.PromoCode{IsActive: false}
	if err := service.ValidatePromo(p); err == nil {
		t.Error("expected error for inactive promo")
	}
}

func TestValidatePromo_ExhaustedActivations(t *testing.T) {
	max := 5
	p := &model.PromoCode{IsActive: true, MaxActivations: &max, ActivationsCount: 5}
	if err := service.ValidatePromo(p); err == nil {
		t.Error("expected error for exhausted activations")
	}
}

func TestValidatePromo_Valid(t *testing.T) {
	future := time.Now().Add(time.Hour)
	max := 10
	p := &model.PromoCode{IsActive: true, ExpiresAt: &future, MaxActivations: &max, ActivationsCount: 3}
	if err := service.ValidatePromo(p); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}
