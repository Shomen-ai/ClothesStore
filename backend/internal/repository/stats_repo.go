package repository

import (
	"database/sql"
	"fmt"
)

type StatsRepo struct{ db *sql.DB }

func NewStatsRepo(db *sql.DB) *StatsRepo { return &StatsRepo{db: db} }

type RevenuePoint struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
}

type OrderStatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type TopProduct struct {
	ProductID int64  `json:"product_id"`
	Name      string `json:"name"`
	Sold      int    `json:"sold"`
}

type PromoStat struct {
	Code        string `json:"code"`
	Activations int    `json:"activations"`
}

func (r *StatsRepo) periodCondition(period string) string {
	switch period {
	case "day":
		return "AND created_at >= CURRENT_DATE"
	case "week":
		return "AND created_at >= CURRENT_DATE - INTERVAL '7 days'"
	case "month":
		return "AND created_at >= date_trunc('month', CURRENT_DATE)"
	case "quarter":
		return "AND created_at >= date_trunc('quarter', CURRENT_DATE)"
	}
	return ""
}

func (r *StatsRepo) GetRevenue(period string) ([]RevenuePoint, error) {
	cond := r.periodCondition(period)
	groupFmt := "YYYY-MM-DD"
	if period == "quarter" || period == "all" {
		groupFmt = "IYYY-IW"
	}
	rows, err := r.db.Query(fmt.Sprintf(
		`SELECT TO_CHAR(created_at,'%s') dt, SUM(total_price)
		 FROM orders WHERE status != 'cancelled' %s
		 GROUP BY TO_CHAR(created_at,'%s') ORDER BY dt`, groupFmt, cond, groupFmt,
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

func (r *StatsRepo) GetOrderCounts(period string) ([]OrderStatusCount, error) {
	cond := r.periodCondition(period)
	rows, err := r.db.Query(fmt.Sprintf(
		`SELECT status, COUNT(*) FROM orders WHERE 1=1 %s GROUP BY status`, cond,
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
