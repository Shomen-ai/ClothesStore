package middleware

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// AdminRequired returns middleware that allows only admin users through. It must
// be chained AFTER AuthRequired, since it relies on the role that AuthRequired
// places in the context under CtxKeyUserRole. A missing role means the request
// was not authenticated (401); a present but non-admin role is rejected (403).
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(CtxKeyUserRole)
		// No role in context => AuthRequired did not run / did not authenticate.
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		// Authenticated but not an admin: forbidden.
		if roleVal != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
