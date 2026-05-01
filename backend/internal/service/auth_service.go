package service

import (
	"database/sql"
	"errors"
	"clothes-store/internal/model"
	"clothes-store/internal/repository"
	appjwt "clothes-store/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

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
	u := &model.User{Email: email, PasswordHash: string(hash), Name: name, Phone: phone, Role: "customer"}
	if err := s.userRepo.Create(u); err != nil {
		return nil, nil, err
	}
	pair, err := s.generatePair(u.ID, u.Role)
	return u, pair, err
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
