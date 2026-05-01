package jwt_test

import (
	"testing"
	"time"
	appjwt "clothes-store/pkg/jwt"
)

const secret = "test-secret-32-characters-long!!"

func TestGenerateAndValidateAccessToken(t *testing.T) {
	token, err := appjwt.GenerateAccessToken(42, "customer", secret)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	claims, err := appjwt.ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("want UserID=42, got %d", claims.UserID)
	}
	if claims.Role != "customer" {
		t.Errorf("want Role=customer, got %s", claims.Role)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, _ := appjwt.GenerateAccessToken(1, "customer", secret)
	_, err := appjwt.ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Error("expected error with wrong secret")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token, err := appjwt.GenerateRefreshToken(99, secret)
	if err != nil {
		t.Fatalf("generate refresh: %v", err)
	}
	claims, err := appjwt.ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("validate refresh: %v", err)
	}
	if claims.UserID != 99 {
		t.Errorf("want UserID=99, got %d", claims.UserID)
	}
	// Refresh token expires in 30 days
	if time.Until(claims.ExpiresAt.Time) < 29*24*time.Hour {
		t.Error("refresh token expiry too short")
	}
}
