package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"clothes-store/internal/repository"

	"github.com/gin-gonic/gin"
)

// admin_review_handler.go serves the admin review-moderation endpoints under
// /api/admin/reviews. All routes sit behind AuthRequired + AdminRequired.
type AdminReviewHandler struct{ repo *repository.ReviewRepo }

// NewAdminReviewHandler constructs an AdminReviewHandler backed by the review repo.
func NewAdminReviewHandler(repo *repository.ReviewRepo) *AdminReviewHandler {
	return &AdminReviewHandler{repo: repo}
}

// List serves GET /api/admin/reviews. The optional ?hidden=true|false (or 1)
// query param filters by visibility; when absent, all reviews are returned. 200
// on success, 500 on a repo error.
func (h *AdminReviewHandler) List(c *gin.Context) {
	var hidden *bool
	if v := c.Query("hidden"); v != "" {
		b := v == "true" || v == "1"
		hidden = &b
	}
	list, err := h.repo.AdminList(hidden)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// adminPatchBody is the body for Patch. IsHidden is a pointer so the handler can
// distinguish an omitted field (nil -> 400) from an explicit true/false.
type adminPatchBody struct {
	IsHidden *bool `json:"is_hidden"`
}

// Patch serves PATCH /api/admin/reviews/:id, toggling a review's visibility.
// Requires a body with is_hidden present. 204 on success, 400 on a bad id or
// missing is_hidden, 404 if no such review, 500 on a repo error.
func (h *AdminReviewHandler) Patch(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body adminPatchBody
	if err := c.ShouldBindJSON(&body); err != nil || body.IsHidden == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "is_hidden required"})
		return
	}
	if err := h.repo.AdminSetHidden(id, *body.IsHidden); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Delete serves DELETE /api/admin/reviews/:id, hard-deleting any review (admin
// is not scoped to ownership). 204 on success, 400 on a bad id, 404 if no such
// review, 500 on a repo error.
func (h *AdminReviewHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.repo.AdminDelete(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
