package service

import (
	"database/sql"
	"errors"
	"strings"
	"clothes-store/internal/model"
	"clothes-store/internal/repository"
	appjwt "clothes-store/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

// ErrEmailTaken is returned when a registration targets an email that already
// has an account; callers translate it into an HTTP 409.
var ErrEmailTaken = errors.New("email already registered")

// AuthService handles user registration, login and JWT issuance. It owns the
// user repository and the secret used to sign access/refresh tokens.
type AuthService struct {
	userRepo  *repository.UserRepo
	jwtSecret string
}

// NewAuthService wires the auth service to its user repository and JWT secret.
func NewAuthService(userRepo *repository.UserRepo, secret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: secret}
}

// TokenPair is the access/refresh JWT pair returned on successful auth.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Register hashes the plaintext password with bcrypt and creates the user,
// returning a fresh token pair. This is the direct (non-email-verified) path.
func (s *AuthService) Register(email, password, name, phone string) (*model.User, *TokenPair, error) {
	// bcrypt salts and hashes the password; the plaintext is never persisted.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, nil, err
	}
	return s.RegisterFromPrehashed(email, string(hash), name, phone)
}

// RegisterFromPrehashed creates a user given an already-bcrypted password.
// Used by the email-verified sign-up path, where the hash lives on the pending
// row from the moment register/start hit.
func (s *AuthService) RegisterFromPrehashed(email, passwordHash, name, phone string) (*model.User, *TokenPair, error) {
	// New users always start with the "customer" role; admins are provisioned
	// out of band (seed/DB), never through this path.
	u := &model.User{Email: email, PasswordHash: passwordHash, Name: name, Phone: phone, Role: "customer"}
	if err := s.userRepo.Create(u); err != nil {
		// A unique-violation on email means the account was created concurrently
		// (or the pre-check raced); surface it as the typed ErrEmailTaken.
		if isUniqueEmailErr(err) {
			return nil, nil, ErrEmailTaken
		}
		return nil, nil, err
	}
	pair, err := s.generatePair(u.ID, u.Role)
	return u, pair, err
}

// EmailTaken is a cheap pre-check so register/start can return 409 before we
// generate a code, hash a password and burn an SMTP call.
func (s *AuthService) EmailTaken(email string) (bool, error) {
	_, err := s.userRepo.GetByEmail(email)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// isUniqueEmailErr reports whether err is a unique-constraint violation on the
// email column. It matches by driver-specific markers (Postgres SQLSTATE 23505,
// the generic "duplicate key" text, or Oracle's ORA-00001) instead of a typed
// error so it stays portable across DB backends.
func isUniqueEmailErr(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "23505") ||
		strings.Contains(s, "duplicate key") ||
		strings.Contains(s, "ORA-00001")
}

// Login verifies the email/password pair and returns a token pair on success.
// A missing user and a wrong password both yield the same opaque "invalid
// credentials" error so the response can't be used to probe which emails exist.
func (s *AuthService) Login(email, password string) (*model.User, *TokenPair, error) {
	u, err := s.userRepo.GetByEmail(email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, errors.New("invalid credentials")
	}
	if err != nil {
		return nil, nil, err
	}
	// bcrypt compares the stored hash against the candidate in constant time.
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, nil, errors.New("invalid credentials")
	}
	pair, err := s.generatePair(u.ID, u.Role)
	return u, pair, err
}

// RefreshToken validates a refresh JWT and, if the referenced user still
// exists, mints a brand-new access/refresh pair (token rotation).
func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	claims, err := appjwt.ValidateToken(refreshToken, s.jwtSecret)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}
	u, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, err
	}
	return s.generatePair(u.ID, u.Role)
}

// generatePair signs a short-lived access token (carrying the role for
// authorization) and a longer-lived refresh token for the given user.
func (s *AuthService) generatePair(userID int64, role string) (*TokenPair, error) {
	access, err := appjwt.GenerateAccessToken(userID, role, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	refresh, err := appjwt.GenerateRefreshToken(userID, s.jwtSecret)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: refresh}, nil
}
