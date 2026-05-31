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

var ErrEmailTaken = errors.New("email already registered")

type AuthService struct {
	userRepo  *repository.UserRepo
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepo, secret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: secret}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *AuthService) Register(email, password, name, phone string) (*model.User, *TokenPair, error) {
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
	u := &model.User{Email: email, PasswordHash: passwordHash, Name: name, Phone: phone, Role: "customer"}
	if err := s.userRepo.Create(u); err != nil {
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

func isUniqueEmailErr(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "23505") ||
		strings.Contains(s, "duplicate key") ||
		strings.Contains(s, "ORA-00001")
}

func (s *AuthService) Login(email, password string) (*model.User, *TokenPair, error) {
	u, err := s.userRepo.GetByEmail(email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, errors.New("invalid credentials")
	}
	if err != nil {
		return nil, nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, nil, errors.New("invalid credentials")
	}
	pair, err := s.generatePair(u.ID, u.Role)
	return u, pair, err
}

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
