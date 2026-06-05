package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"clothes-store/internal/middleware"
	appjwt "clothes-store/pkg/jwt"
	"github.com/gin-gonic/gin"
)

const secret = "test-secret-32-characters-long!!"

// setupRouter builds a test engine with a customer-protected route and an
// admin-only route, both guarded by the real middleware chain.
func setupRouter(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", middleware.AuthRequired(secret), func(c *gin.Context) {
		id := c.GetInt64("userID")
		c.JSON(200, gin.H{"user_id": id})
	})
	r.GET("/admin", middleware.AuthRequired(secret), middleware.AdminRequired(), func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})
	return r
}

// TestAuthRequired_NoToken: a request with no Authorization header is rejected 401.
func TestAuthRequired_NoToken(t *testing.T) {
	r := setupRouter(secret)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	r.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Errorf("want 401, got %d", w.Code)
	}
}

// TestAuthRequired_ValidToken: a valid Bearer token passes the guard with 200.
func TestAuthRequired_ValidToken(t *testing.T) {
	r := setupRouter(secret)
	token, _ := appjwt.GenerateAccessToken(7, "customer", secret)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("want 200, got %d", w.Code)
	}
}

// TestAdminRequired_CustomerForbidden: a customer token authenticates but is
// blocked from the admin route with 403.
func TestAdminRequired_CustomerForbidden(t *testing.T) {
	r := setupRouter(secret)
	token, _ := appjwt.GenerateAccessToken(7, "customer", secret)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Errorf("want 403, got %d", w.Code)
	}
}
