package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"clothes-store/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type ReviewHandler struct{ repo *repository.ReviewRepo }

func NewReviewHandler(repo *repository.ReviewRepo) *ReviewHandler {
	return &ReviewHandler{repo: repo}
}

const maxReviewTextLen = 2000

// GET /api/products/:id/reviews (public)
func (h *ReviewHandler) ListForProduct(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	list, err := h.repo.ListForProduct(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	sum, err := h.repo.SummaryForProduct(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"reviews":    list,
		"avg_rating": sum.AvgRating,
		"count":      sum.Count,
	})
}

// GET /api/products/:id/reviews/me (auth)
// Returns {eligible: bool, own: Review|null} so the UI can render the right form state.
func (h *ReviewHandler) MyForProduct(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	userID := c.GetInt64("userID")

	eligible, err := h.repo.IsEligible(userID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	own, err := h.repo.GetMine(userID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"eligible": eligible, "own": own})
}

type reviewBody struct {
	Rating int    `json:"rating"`
	Text   string `json:"text"`
}

func (b *reviewBody) validate() string {
	if b.Rating < 1 || b.Rating > 5 {
		return "rating must be between 1 and 5"
	}
	b.Text = strings.TrimSpace(b.Text)
	if len(b.Text) > maxReviewTextLen {
		return "text too long"
	}
	return ""
}

// POST /api/products/:id/reviews (auth)
func (h *ReviewHandler) Create(c *gin.Context) {
	productID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	userID := c.GetInt64("userID")

	var body reviewBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if msg := body.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	ok, err := h.repo.IsEligible(userID, productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "отзыв доступен после получения товара"})
		return
	}

	rv, err := h.repo.Create(userID, productID, body.Rating, body.Text)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "отзыв уже оставлен"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rv)
}

// PUT /api/products/:id/reviews/:rid (auth, own)
func (h *ReviewHandler) Update(c *gin.Context) {
	reviewID, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil || reviewID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}
	userID := c.GetInt64("userID")

	var body reviewBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if msg := body.validate(); msg != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	rv, err := h.repo.Update(reviewID, userID, body.Rating, body.Text)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "отзыв не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rv)
}

// DELETE /api/products/:id/reviews/:rid (auth, own)
func (h *ReviewHandler) Delete(c *gin.Context) {
	reviewID, err := strconv.ParseInt(c.Param("rid"), 10, 64)
	if err != nil || reviewID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return
	}
	userID := c.GetInt64("userID")

	if err := h.repo.Delete(reviewID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "отзыв не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
