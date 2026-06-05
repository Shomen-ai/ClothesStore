package jwt

import (
	"fmt"
	"time"
	"github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT payload carried by both access and refresh tokens. It embeds
// the standard RegisteredClaims (exp, iat, ...) and adds the app-specific user
// identity. Role is set only on access tokens; refresh tokens leave it empty.
type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateAccessToken mints a short-lived (15 minute) HS256 access token carrying
// the user's ID and role. The role is what AdminRequired later checks. Returns an
// error if secret is empty.
func GenerateAccessToken(userID int64, role, secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("jwt: secret must not be empty")
	}
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)), // short access lifetime
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken mints a long-lived (30 day) HS256 refresh token. It only
// identifies the user (no role) and is exchanged for fresh access tokens.
// Returns an error if secret is empty.
func GenerateRefreshToken(userID int64, secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("jwt: secret must not be empty")
	}
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)), // long refresh lifetime (30 days)
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ValidateToken parses and verifies a token string against secret and returns its
// Claims. It rejects an empty secret, and crucially enforces that the token was
// signed with an HMAC method: the keyfunc asserts t.Method is *SigningMethodHMAC
// before handing back the secret. This blocks the classic "alg confusion" attack
// where an attacker swaps the algorithm (e.g. to "none" or RS256) to bypass the
// signature check. Expired, tampered, or otherwise invalid tokens return an error.
func ValidateToken(tokenStr, secret string) (*Claims, error) {
	if secret == "" {
		return nil, fmt.Errorf("jwt: secret must not be empty")
	}
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		// Enforce HMAC signing only; reject any other alg to prevent alg confusion.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	// Final guard: claims type assertion plus the library's own validity flag.
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}
