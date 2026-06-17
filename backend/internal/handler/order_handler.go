package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"clothes-store/internal/repository"
	"clothes-store/internal/service"
	"github.com/gin-gonic/gin"
)

// order_handler.go serves the authenticated customer order endpoints under /api
// (order creation, listing, payment confirmation and promo validation). All
// routes sit behind AuthRequired; the user id is read from the JWT via
// c.GetInt64("userID"), never from the request body.
type OrderHandler struct {
	svc       *service.OrderService
	promoRepo *repository.PromoRepo
	orderRepo *repository.OrderRepo
}

// NewOrderHandler wires the order service plus the promo and order repos.
func NewOrderHandler(svc *service.OrderService, pr *repository.PromoRepo, or *repository.OrderRepo) *OrderHandler {
	return &OrderHandler{svc: svc, promoRepo: pr, orderRepo: or}
}

// Create serves POST /api/orders, placing an order for the authenticated user.
// The service re-prices each item from the DB, validates and applies any promo,
// decrements stock and persists the order in one transaction. 201 with the
// created order, 400 on a bad body or any service error (e.g. empty cart,
// unavailable item, invalid promo).
func (h *OrderHandler) Create(c *gin.Context) {
	userID := c.GetInt64("userID")
	var req service.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	order, err := h.svc.Create(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

// GetUserOrders serves GET /api/user/orders, returning the authenticated user's
// orders newest first. 200 on success, 500 on a repo error.
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID := c.GetInt64("userID")
	orders, err := h.orderRepo.GetByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// GetUserOrder serves GET /api/user/orders/:id. Ownership is enforced: an order
// belonging to another user returns 404 (not 403) to avoid leaking existence.
// 200 on success, 400 on a bad id, 404 if missing or not owned, 500 on a repo
// error.
func (h *OrderHandler) GetUserOrder(c *gin.Context) {
	userID := c.GetInt64("userID")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	o, err := h.orderRepo.GetByID(id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if o.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, o)
}

// ConfirmPayment marks an order as paid. Used by the payment-stub page after
// the user "completes" payment on the fake checkout form. Idempotent: a
// second call on an already-confirmed order returns 200 silently.
func (h *OrderHandler) ConfirmPayment(c *gin.Context) {
	userID := c.GetInt64("userID")
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}
	o, err := h.orderRepo.GetByID(id)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if o.UserID != userID {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	// Status transition guard: only a pending order moves to confirmed; terminal
	// or already-advanced states are handled idempotently or rejected.
	switch o.Status {
	case "confirmed", "shipped", "delivered":
		// Already past pending — idempotent success.
		c.JSON(http.StatusOK, gin.H{"status": o.Status})
		return
	case "cancelled":
		c.JSON(http.StatusConflict, gin.H{"error": "заказ отменён"})
		return
	case "pending":
		if err := h.orderRepo.ConfirmPaid(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "confirmed"})
		return
	default:
		c.JSON(http.StatusConflict, gin.H{"error": "недопустимый статус заказа"})
	}
}

// ValidatePromo serves POST /api/promo/validate, a preview used by checkout
// before placing an order. It looks up the code, validates it (active, not
// expired, under its activation cap) and computes the discount for the supplied
// total without mutating anything. 200 with the discount breakdown, 400 on a bad
// body or an invalid/expired promo, 404 if the code does not exist. Note: Total
// uses binding:"required", so a legitimate total of 0 is rejected (see findings).
func (h *OrderHandler) ValidatePromo(c *gin.Context) {
	var req struct {
		Code  string  `json:"code" binding:"required"`
		Total float64 `json:"total" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	promo, err := h.promoRepo.GetByCode(req.Code)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "promo not found"})
		return
	}
	if err := service.ValidatePromo(promo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	discount := service.ApplyDiscount(promo, req.Total)
	c.JSON(http.StatusOK, gin.H{
		"valid":           true,
		"discount_type":   promo.DiscountType,
		"discount_value":  promo.DiscountValue,
		"discount_amount": discount,
		"final_total":     req.Total - discount,
	})
}
