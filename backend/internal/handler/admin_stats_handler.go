package handler

import (
	"net/http"
	"clothes-store/internal/repository"
	"github.com/gin-gonic/gin"
)

// admin_stats_handler.go serves the admin dashboard statistics endpoints under
// /api/admin/stats. All routes sit behind AuthRequired + AdminRequired.
type AdminStatsHandler struct{ repo *repository.StatsRepo }

// NewAdminStatsHandler constructs an AdminStatsHandler backed by the stats repo.
func NewAdminStatsHandler(repo *repository.StatsRepo) *AdminStatsHandler {
	return &AdminStatsHandler{repo: repo}
}

// validPeriods is the allowed set for the ?period query param shared by every
// stats endpoint. The repo maps each value to a SQL date window; passing the
// constant rather than the raw param keeps it out of the query as text.
var validPeriods = map[string]bool{"day": true, "week": true, "month": true, "quarter": true, "all": true}

// Revenue serves GET /api/admin/stats/revenue. Period defaults to "month" and
// must be in validPeriods. 200 with the revenue series, 400 on an invalid
// period, 500 on a repo error.
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

// Orders serves GET /api/admin/stats/orders, returning order counts by status
// plus the top products for the period. Period defaults to "month" and must be
// in validPeriods. 200 with {by_status, top_products}, 400 on an invalid period,
// 500 on a repo error.
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

// Promos serves GET /api/admin/stats/promo-codes, returning promo-code usage
// stats. Period defaults to "month" and must be in validPeriods. 200 on success,
// 400 on an invalid period, 500 on a repo error.
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
