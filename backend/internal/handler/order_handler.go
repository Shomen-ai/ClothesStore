package handler

import (
	"database/sql"
	"net/http"
	"strconv"
	"clothes-store/internal/repository"
	"clothes-store/internal/service"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc       *service.OrderService
	promoRepo *repository.PromoRepo
	orderRepo *repository.OrderRepo
}

func NewOrderHandler(svc *service.OrderService, pr *repository.PromoRepo, or *repository.OrderRepo) *OrderHandler {
	return &OrderHandler{svc: svc, promoRepo: pr, orderRepo: or}
}

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

func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID := c.GetInt64("userID")
	orders, err := h.orderRepo.GetByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

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
	switch o.Status {
	case "confirmed", "shipped", "delivered":
		// Already past pending — idempotent success.
		c.JSON(http.StatusOK, gin.H{"status": o.Status})
		return
	case "cancelled":
		c.JSON(http.StatusConflict, gin.H{"error": "заказ отменён"})
		return
	case "pending":
		if err := h.orderRepo.UpdateStatus(id, "confirmed"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "confirmed"})
		return
	default:
		c.JSON(http.StatusConflict, gin.H{"error": "недопустимый статус заказа"})
	}
}

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
