package handler

import (
	"net/http"
	"strconv"
	"clothes-store/internal/repository"
	"github.com/gin-gonic/gin"
)

type WishlistHandler struct{ repo *repository.WishlistRepo }

func NewWishlistHandler(repo *repository.WishlistRepo) *WishlistHandler {
	return &WishlistHandler{repo: repo}
}

func (h *WishlistHandler) Get(c *gin.Context) {
	userID := c.GetInt64("userID")
	items, err := h.repo.Get(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *WishlistHandler) Add(c *gin.Context) {
	userID := c.GetInt64("userID")
	productID, err := strconv.ParseInt(c.Param("product_id"), 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	if err := h.repo.Add(userID, productID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "already in wishlist"})
		return
	}
	c.Status(http.StatusCreated)
}

func (h *WishlistHandler) Remove(c *gin.Context) {
	userID := c.GetInt64("userID")
	productID, err := strconv.ParseInt(c.Param("product_id"), 10, 64)
	if err != nil || productID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}
	if err := h.repo.Remove(userID, productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
