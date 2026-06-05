// Package repository: data-access layer for admin-dashboard statistics.
package repository

import (
	"database/sql"
	"fmt"
)

// StatsRepo provides the aggregated read-only queries backing the admin
// dashboard charts (revenue, order status mix, top products, promo usage).
type StatsRepo struct{ db *sql.DB }

// NewStatsRepo constructs a StatsRepo bound to db.
func NewStatsRepo(db *sql.DB) *StatsRepo { return &StatsRepo{db: db} }

// RevenuePoint is one bucket (day or week) of summed revenue for the chart.
type RevenuePoint struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
}

// OrderStatusCount is the order tally for a single status within the period.
type OrderStatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// TopProduct is one best-seller row: product plus total units sold in the period.
type TopProduct struct {
	ProductID int64  `json:"product_id"`
	Name      string `json:"name"`
	Sold      int    `json:"sold"`
}

// PromoStat is one promo code with its activation count for the period.
type PromoStat struct {
	Code        string `json:"code"`
	Activations int    `json:"activations"`
}

// periodCondition is appended to dashboard stats queries. The upper bound is
// load-bearing: the demo seed extends orders into the future so the chart
// stays populated as time moves; without clamping at NOW() (or end-of-today
// for the "day" period) those future orders bleed into every bucket and the
// chart shows revenue for days that haven't happened yet.
func (r *StatsRepo) periodCondition(period string) string {
	switch period {
	case "day":
		return "AND o.created_at >= CURRENT_DATE AND o.created_at < CURRENT_DATE + INTERVAL '1 day'"
	case "week":
		return "AND o.created_at >= CURRENT_DATE - INTERVAL '7 days' AND o.created_at <= NOW()"
	case "month":
		return "AND o.created_at >= date_trunc('month', CURRENT_DATE) AND o.created_at <= NOW()"
	case "quarter":
		return "AND o.created_at >= date_trunc('quarter', CURRENT_DATE) AND o.created_at <= NOW()"
	case "all":
		return "AND o.created_at <= NOW()"
	}
	return ""
}

// GetRevenue returns the revenue time series for the period, excluding cancelled
// orders. Buckets are per-day for short ranges, but switch to per-week
// (date_trunc('week', …)) for the wide "quarter" and "all" ranges to keep the
// chart readable. The bucket expression is interpolated into GROUP BY/ORDER BY
// via fmt.Sprintf — it is a fixed internal string, not user input.
func (r *StatsRepo) GetRevenue(period string) ([]RevenuePoint, error) {
	cond := r.periodCondition(period)
	groupExpr := "o.created_at::date"
	if period == "quarter" || period == "all" {
		groupExpr = "date_trunc('week', o.created_at)::date"
	}
	rows, err := r.db.Query(fmt.Sprintf(
		`SELECT TO_CHAR(%s, 'DD.MM.YYYY') AS dt, SUM(o.total_price)
		 FROM orders o WHERE o.status != 'cancelled' %s
		 GROUP BY %s ORDER BY %s`, groupExpr, cond, groupExpr, groupExpr,
	))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	points := make([]RevenuePoint, 0)
	for rows.Next() {
		var p RevenuePoint
		if err := rows.Scan(&p.Date, &p.Revenue); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return points, nil
}

// GetOrderCounts returns the order count per status for the period (all statuses
// included — the status mix donut). The WHERE 1=1 seed lets the period clause
// be appended unconditionally.
func (r *StatsRepo) GetOrderCounts(period string) ([]OrderStatusCount, error) {
	cond := r.periodCondition(period)
	rows, err := r.db.Query(fmt.Sprintf(
		`SELECT o.status, COUNT(*) FROM orders o WHERE 1=1 %s GROUP BY o.status`, cond,
	))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	counts := make([]OrderStatusCount, 0)
	for rows.Next() {
		var c OrderStatusCount
		if err := rows.Scan(&c.Status, &c.Count); err != nil {
			return nil, err
		}
		counts = append(counts, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return counts, nil
}

// GetTopProducts returns the 5 best-selling products (by units) in the period,
// joining line items to their order (for the date/status filter) and product (for the name).
func (r *StatsRepo) GetTopProducts(period string) ([]TopProduct, error) {
	cond := r.periodCondition(period)
	rows, err := r.db.Query(fmt.Sprintf(
		`SELECT oi.product_id, p.name, SUM(oi.quantity) sold
		 FROM order_items oi
		 JOIN orders o ON o.id=oi.order_id
		 JOIN products p ON p.id=oi.product_id
		 WHERE o.status != 'cancelled' %s
		 GROUP BY oi.product_id, p.name
		 ORDER BY sold DESC LIMIT 5`, cond,
	))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	top := make([]TopProduct, 0)
	for rows.Next() {
		var t TopProduct
		if err := rows.Scan(&t.ProductID, &t.Name, &t.Sold); err != nil {
			return nil, err
		}
		top = append(top, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return top, nil
}

// GetPromoStats returns promo activation counts for the period. The "all" case
// reads the precomputed activations_count column directly (cheap, all-time); any
// scoped period instead counts distinct orders that used each code within the
// window (ORDER BY 2 = order by the count column).
func (r *StatsRepo) GetPromoStats(period string) ([]PromoStat, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if period == "all" {
		rows, err = r.db.Query(
			`SELECT code, activations_count FROM promo_codes
			 WHERE activations_count > 0 ORDER BY activations_count DESC`,
		)
	} else {
		cond := r.periodCondition(period)
		rows, err = r.db.Query(fmt.Sprintf(
			`SELECT pc.code, COUNT(DISTINCT o.id)
			 FROM promo_codes pc
			 JOIN orders o ON o.promo_code_id=pc.id
			 WHERE 1=1 %s
			 GROUP BY pc.code ORDER BY 2 DESC`, cond,
		))
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	stats := make([]PromoStat, 0)
	for rows.Next() {
		var s PromoStat
		if err := rows.Scan(&s.Code, &s.Activations); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stats, nil
}
