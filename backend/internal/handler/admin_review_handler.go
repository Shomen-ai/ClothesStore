package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"clothes-store/internal/repository"

	"github.com/gin-gonic/gin"
)

type AdminReviewHandler struct{ repo *repository.ReviewRepo }

func NewAdminReviewHandler(repo *repository.ReviewRepo) *AdminReviewHandler {
	return &AdminReviewHandler{repo: repo}
}

// GET /api/admin/reviews?hidden=true|false
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

type adminPatchBody struct {
	IsHidden *bool `json:"is_hidden"`
}

// PATCH /api/admin/reviews/:id
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

// DELETE /api/admin/reviews/:id
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
