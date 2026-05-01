package middleware

import (
	"strings"
	"net/http"
	appjwt "clothes-store/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := appjwt.ValidateToken(tokenStr, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyUserRole, claims.Role)
		c.Next()
	}
}
