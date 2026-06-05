package middleware

import (
	"strings"
	"net/http"
	appjwt "clothes-store/pkg/jwt"
	"github.com/gin-gonic/gin"
)

// AuthRequired returns middleware that authenticates the request via a Bearer
// JWT. It reads the Authorization header, requires the "Bearer " prefix, and
// validates the token against secret using appjwt.ValidateToken. Any missing
// header, wrong prefix, or validation failure aborts the request with 401. On
// success it stashes the authenticated user's ID and role in the gin context
// under CtxKeyUserID / CtxKeyUserRole so downstream handlers (and AdminRequired)
// can read them.
func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		// Reject anything that isn't a Bearer token (also catches an empty header).
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		// Strip the scheme prefix to get the raw token string.
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		// Verify signature and claims; rejects expired/tampered/wrong-secret tokens.
		claims, err := appjwt.ValidateToken(tokenStr, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		// Expose identity to downstream handlers and the AdminRequired gate.
		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyUserRole, claims.Role)
		c.Next()
	}
}
