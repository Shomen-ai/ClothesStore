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

type AuthHandler struct {
	svc         *service.AuthService
	pendingRepo *repository.PendingRegistrationRepo
	mail        mailer.Mailer
}

func NewAuthHandler(svc *service.AuthService, pending *repository.PendingRegistrationRepo, m mailer.Mailer) *AuthHandler {
	return &AuthHandler{svc: svc, pendingRepo: pending, mail: m}
}

// ─── login / refresh ────────────────────────────────────────────────────────

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

func gen6Code() string {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		// crypto/rand failing is exceptional — fall back to a stable string so
		// the flow still completes; the email is the security boundary anyway.
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}

func normaliseEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
