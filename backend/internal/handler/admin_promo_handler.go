package handler

import (
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"clothes-store/internal/model"
	"clothes-store/internal/repository"
	"github.com/gin-gonic/gin"
)

type AdminPromoHandler struct{ repo *repository.PromoRepo }

func NewAdminPromoHandler(repo *repository.PromoRepo) *AdminPromoHandler {
	return &AdminPromoHandler{repo: repo}
}

func (h *AdminPromoHandler) List(c *gin.Context) {
	promos, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, promos)
}

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

func (h *AdminPromoHandler) Deactivate(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Deactivate(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminPromoHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func randomCode(n int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := strings.Builder{}
	for i := 0; i < n; i++ {
		b.WriteByte(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
