package handler

import (
	"net/http"
	"clothes-store/internal/repository"
	"github.com/gin-gonic/gin"
)

type AdminStatsHandler struct{ repo *repository.StatsRepo }

func NewAdminStatsHandler(repo *repository.StatsRepo) *AdminStatsHandler {
	return &AdminStatsHandler{repo: repo}
}

var validPeriods = map[string]bool{"day": true, "week": true, "month": true, "quarter": true, "all": true}

func (h *AdminStatsHandler) Revenue(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	if !validPeriods[period] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period"})
		return
	}
	data, err := h.repo.GetRevenue(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *AdminStatsHandler) Orders(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	if !validPeriods[period] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period"})
		return
	}
	data, err := h.repo.GetOrderCounts(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	top, err := h.repo.GetTopProducts(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"by_status": data, "top_products": top})
}

func (h *AdminStatsHandler) Promos(c *gin.Context) {
	period := c.DefaultQuery("period", "month")
	if !validPeriods[period] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period"})
		return
	}
	data, err := h.repo.GetPromoStats(period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
