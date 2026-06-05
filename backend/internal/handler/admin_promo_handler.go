package handler

import (
	"crypto/rand"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"clothes-store/internal/model"
	"clothes-store/internal/repository"
	"github.com/gin-gonic/gin"
)

// AdminPromoHandler serves the admin promo-code endpoints under
// /api/admin/promo-codes. All routes sit behind AuthRequired + AdminRequired.
type AdminPromoHandler struct{ repo *repository.PromoRepo }

// NewAdminPromoHandler constructs an AdminPromoHandler backed by the given repo.
func NewAdminPromoHandler(repo *repository.PromoRepo) *AdminPromoHandler {
	return &AdminPromoHandler{repo: repo}
}

// List serves GET /api/admin/promo-codes, returning all promo codes newest
// first. 200 on success, 500 on a repo error.
func (h *AdminPromoHandler) List(c *gin.Context) {
	promos, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, promos)
}

// Create serves POST /api/admin/promo-codes. If the body omits a code, a random
// 8-character code is generated. The promo is always created active. 201 with
// the created promo, 400 on a bad body, 500 on a repo error.
func (h *AdminPromoHandler) Create(c *gin.Context) {
	var p model.PromoCode
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if p.Code == "" {
		p.Code = randomCode(8)
	}
	p.IsActive = true
	if err := h.repo.Create(&p); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, p)
}

// Deactivate serves PUT /api/admin/promo-codes/:id/deactivate, setting
// is_active=false. 204 on success, 500 on a repo error.
func (h *AdminPromoHandler) Deactivate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Deactivate(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Delete serves DELETE /api/admin/promo-codes/:id. 204 on success, 500 on a repo
// error.
func (h *AdminPromoHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// randomCode returns an n-character code over an uppercase-alphanumeric alphabet
// using crypto/rand. The rand.Int error is ignored; on the (practically
// impossible) failure path idx is nil and the call would panic.
func randomCode(n int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := strings.Builder{}
	for range n {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b.WriteByte(chars[idx.Int64()])
	}
	return b.String()
}
