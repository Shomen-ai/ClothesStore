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

// review_handler.go serves the product-review endpoints. ListForProduct is
// public; the rest sit behind AuthRequired. Write operations are authorised
// against the JWT user id (c.GetInt64("userID")) and, for posting, against
// purchase eligibility.
type ReviewHandler struct{ repo *repository.ReviewRepo }

// NewReviewHandler constructs a ReviewHandler backed by the review repo.
func NewReviewHandler(repo *repository.ReviewRepo) *ReviewHandler {
	return &ReviewHandler{repo: repo}
}

// maxReviewTextLen caps the review body length accepted by validate().
const maxReviewTextLen = 2000

// ListForProduct serves GET /api/products/:id/reviews (public). Returns visible
// reviews plus the aggregate {avg_rating, count}. 200 on success, 400 on a bad
// product id, 500 on a repo error.
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

// MyForProduct serves GET /api/products/:id/reviews/me (auth). It reports whether
// the user may review the product (has a delivered order containing it) and
// returns their existing review if any, as {eligible, own}, so the UI can pick
// the right form state. 200 on success, 400 on a bad product id, 500 on a repo
// error.
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

// reviewBody is the shared request body for Create and Update.
type reviewBody struct {
	Rating int    `json:"rating"`
	Text   string `json:"text"`
}

// validate checks the rating is 1..5 and trims the text in place, rejecting text
// longer than maxReviewTextLen. It returns an empty string when valid, otherwise
// a client-facing error message.
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

// Create serves POST /api/products/:id/reviews (auth). It validates the body,
// then requires the user to be eligible (a delivered order containing the
// product) -> 403 otherwise. The unique (user_id, product_id) constraint means a
// second review maps to 409. 201 with the created review, 400 on a bad
// id/body/validation, 403 if not eligible, 409 if already reviewed, 500
// otherwise.
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

// Update serves PUT /api/products/:id/reviews/:rid (auth). Ownership is enforced
// in SQL (WHERE id AND user_id); a non-owned or missing review yields
// sql.ErrNoRows -> 404. 200 with the updated review, 400 on a bad id/body, 404
// if not found or not owned, 500 otherwise.
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

// Delete serves DELETE /api/products/:id/reviews/:rid (auth). Ownership is
// enforced in SQL; a non-owned or missing review yields 404. 204 on success, 400
// on a bad id, 404 if not found or not owned, 500 otherwise.
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
