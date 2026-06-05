// admin_order_handler.go serves the admin order-management endpoints under
// /api/admin/orders. Every route here sits behind AuthRequired + AdminRequired,
// so handlers assume the caller is an authenticated admin.
package handler

import (
	"net/http"
	"strconv"
	"clothes-store/internal/repository"
	"github.com/gin-gonic/gin"
)

// AdminOrderHandler wires the order repository to the admin HTTP routes.
type AdminOrderHandler struct{ repo *repository.OrderRepo }

// NewAdminOrderHandler constructs an AdminOrderHandler backed by the given repo.
func NewAdminOrderHandler(repo *repository.OrderRepo) *AdminOrderHandler {
	return &AdminOrderHandler{repo: repo}
}

// validStatuses is the allowed set of order statuses; UpdateStatus rejects any
// value outside this set before touching the database.
var validStatuses = map[string]bool{
	"pending": true, "confirmed": true, "shipped": true, "delivered": true, "cancelled": true,
}

// List serves GET /api/admin/orders. Optional query params status, date_from and
// date_to filter the result set (forwarded to the repo as parameterised SQL).
// Returns 200 with the matching orders, or 500 on a repo error.
func (h *AdminOrderHandler) List(c *gin.Context) {
	orders, err := h.repo.GetAll(c.Query("status"), c.Query("date_from"), c.Query("date_to"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// Get serves GET /api/admin/orders/:id and returns a single order with its
// items. A missing or unparsable id maps to 404 (an invalid id yields 0, which
// will not be found). Returns 200 on success.
func (h *AdminOrderHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	o, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, o)
}

// UpdateStatus serves PUT /api/admin/orders/:id/status. It requires a JSON body
// {"status": ...} whose value must be one of validStatuses; anything else is
// rejected with 400. Returns 200 with the new status, 400 on bad input, or 500
// on a repo error.
func (h *AdminOrderHandler) UpdateStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !validStatuses[req.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}
	if err := h.repo.UpdateStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": req.Status})
}
