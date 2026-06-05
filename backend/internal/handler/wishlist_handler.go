package handler

import (
	"net/http"
	"strconv"
	"clothes-store/internal/repository"
	"github.com/gin-gonic/gin"
)

// wishlist_handler.go serves the authenticated user's wishlist endpoints under
// /api/user/wishlist. All routes sit behind AuthRequired; the wishlist is keyed
// to the JWT user id (c.GetInt64("userID")).
type WishlistHandler struct{ repo *repository.WishlistRepo }

// NewWishlistHandler constructs a WishlistHandler backed by the wishlist repo.
func NewWishlistHandler(repo *repository.WishlistRepo) *WishlistHandler {
	return &WishlistHandler{repo: repo}
}

// Get serves GET /api/user/wishlist, returning the user's wishlist items. 200 on
// success, 500 on a repo error.
func (h *WishlistHandler) Get(c *gin.Context) {
	userID := c.GetInt64("userID")
	items, err := h.repo.Get(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// Add serves POST /api/user/wishlist/:product_id. 201 on success. Any repo error
// is reported as 409 "already in wishlist"; this assumes the only failure is the
// duplicate-key constraint, so other errors are mislabelled (see findings).
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

// Remove serves DELETE /api/user/wishlist/:product_id. 204 on success (also when
// the item was not present), 400 on a bad product id, 500 on a repo error.
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
