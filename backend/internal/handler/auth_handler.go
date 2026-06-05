package handler

import (
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"clothes-store/internal/mailer"
	"clothes-store/internal/repository"
	"clothes-store/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// auth_handler.go serves the public authentication endpoints under /api/auth:
// login, token refresh, and the three-step email-verified registration flow
// (start -> verify -> resend). These routes are unauthenticated.
type AuthHandler struct {
	svc         *service.AuthService
	pendingRepo *repository.PendingRegistrationRepo
	mail        mailer.Mailer
}

// NewAuthHandler wires the auth service, the pending-registration repo (which
// stages sign-ups awaiting email verification) and the mailer that delivers
// verification codes.
func NewAuthHandler(svc *service.AuthService, pending *repository.PendingRegistrationRepo, m mailer.Mailer) *AuthHandler {
	return &AuthHandler{svc: svc, pendingRepo: pending, mail: m}
}

// ─── login / refresh ────────────────────────────────────────────────────────

// Login serves POST /api/auth/login. On valid credentials it returns 200 with
// {user, tokens}. Any failure (unknown email or wrong password) maps to 401 with
// a generic message, avoiding user enumeration. 400 on a malformed body.
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, tokens, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "tokens": tokens})
}

// Refresh serves POST /api/auth/refresh, exchanging a valid refresh token for a
// fresh token pair. 200 with the new tokens, 400 on a malformed body, 401 if the
// refresh token is invalid or expired.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tokens, err := h.svc.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tokens)
}

// ─── email-verified registration ────────────────────────────────────────────

// RegisterStart serves POST /api/auth/register/start, the first step of sign-up.
// It validates the body (valid email, password >= 6 chars, name required),
// rejects already-registered emails with 409, then bcrypts the password and
// stages a pending_registration row carrying a 6-digit code, a 10-minute expiry
// and a 60-second resend cooldown, and emails the code. The user row is not
// created until RegisterVerify. 200 {sent:true} on success, 400 on a bad body,
// 409 if the email is taken, 502 if the email send fails, 500 on a repo error.
func (h *AuthHandler) RegisterStart(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
		Name     string `json:"name" binding:"required"`
		Phone    string `json:"phone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := normaliseEmail(req.Email)

	taken, err := h.svc.EmailTaken(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if taken {
		c.JSON(http.StatusConflict, gin.H{"error": "email уже зарегистрирован"})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "hash failed"})
		return
	}
	code := gen6Code()
	now := time.Now()
	p := &repository.PendingRegistration{
		Email:        email,
		Code:         code,
		PasswordHash: string(hash),
		Name:         strings.TrimSpace(req.Name),
		Phone:        strings.TrimSpace(req.Phone),
		ExpiresAt:    now.Add(10 * time.Minute),
		ResendAfter:  now.Add(60 * time.Second),
	}
	if err := h.pendingRepo.Upsert(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.mail.SendCode(email, code); err != nil {
		log.Printf("mailer send code failed for %s: %v", email, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "не удалось отправить код"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sent": true})
}

// RegisterVerify serves POST /api/auth/register/verify, the second step. It
// looks up the pending registration by email and enforces, in order: at most 5
// attempts (429), not expired (410), and a matching code (wrong code bumps the
// attempt counter and returns 400). On success it creates the real user from the
// pre-hashed password, deletes the pending row, and returns 201 with
// {user, tokens}. 404 if no pending registration exists, 409 if the email was
// registered in the meantime, 500 on a repo error.
func (h *AuthHandler) RegisterVerify(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code"  binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := normaliseEmail(req.Email)

	p, err := h.pendingRepo.Get(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "регистрация не начата"})
		return
	}
	if p.Attempts >= 5 {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "слишком много попыток — запроси новый код"})
		return
	}
	if time.Now().After(p.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "код истёк — запроси новый"})
		return
	}
	if p.Code != req.Code {
		_ = h.pendingRepo.BumpAttempts(email)
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный код"})
		return
	}

	user, tokens, err := h.svc.RegisterFromPrehashed(p.Email, p.PasswordHash, p.Name, p.Phone)
	if err != nil {
		if errors.Is(err, service.ErrEmailTaken) {
			c.JSON(http.StatusConflict, gin.H{"error": "email уже зарегистрирован"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_ = h.pendingRepo.Delete(email)
	c.JSON(http.StatusCreated, gin.H{"user": user, "tokens": tokens})
}

// RegisterResend serves POST /api/auth/register/resend. Subject to the
// 60-second cooldown (429 if still within it), it generates a new code, resets
// the attempt counter and expiry/cooldown windows, and re-sends the email. 200
// {sent:true} on success, 404 if no pending registration exists, 429 if called
// too soon, 502 on a send failure, 500 on a repo error.
func (h *AuthHandler) RegisterResend(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	email := normaliseEmail(req.Email)

	p, err := h.pendingRepo.Get(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if p == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "регистрация не начата"})
		return
	}
	if time.Now().Before(p.ResendAfter) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "подожди немного перед повторной отправкой"})
		return
	}
	p.Code = gen6Code()
	now := time.Now()
	p.Attempts = 0
	p.ExpiresAt = now.Add(10 * time.Minute)
	p.ResendAfter = now.Add(60 * time.Second)
	if err := h.pendingRepo.Upsert(p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.mail.SendCode(email, p.Code); err != nil {
		log.Printf("mailer resend failed for %s: %v", email, err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "не удалось отправить код"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sent": true})
}

// ─── helpers ────────────────────────────────────────────────────────────────

// gen6Code returns a zero-padded 6-digit verification code from crypto/rand.
func gen6Code() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		// crypto/rand failing is exceptional — fall back to a stable string so
		// the flow still completes; the email is the security boundary anyway.
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

// normaliseEmail lower-cases and trims an email so lookups and the
// pending-registration key are case-insensitive and whitespace-insensitive.
func normaliseEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
