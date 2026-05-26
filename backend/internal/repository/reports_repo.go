package repository

import (
	"database/sql"
)

// ReportsRepo aggregates all read-only queries used by the Excel report.
// Every method returns rows as [][]any so the handler can write them
// straight to xlsx without per-report mapping ceremony.
type ReportsRepo struct{ db *sql.DB }

func NewReportsRepo(db *sql.DB) *ReportsRepo { return &ReportsRepo{db: db} }

type ReportSheet struct {
	Name    string
	Headers []string
	Rows    [][]any
}

func (r *ReportsRepo) queryRows(q string, args ...any) ([][]any, error) {
	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	out := make([][]any, 0)
	for rows.Next() {
		raw := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range raw {
			ptrs[i] = &raw[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		// normalize []byte (pq numeric/text) to string
		for i, v := range raw {
			if b, ok := v.([]byte); ok {
				raw[i] = string(b)
			}
		}
		out = append(out, raw)
	}
	return out, rows.Err()
}

// ────────────────────────────────────────────────────────────────────

func (r *ReportsRepo) Summary() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT
			(SELECT COUNT(*) FROM orders)                                              AS total_orders,
			(SELECT COUNT(*) FROM orders WHERE status='delivered')                     AS delivered,
			(SELECT COUNT(*) FROM orders WHERE status='cancelled')                     AS cancelled,
			(SELECT COALESCE(SUM(total_price),0) FROM orders WHERE status!='cancelled') AS gross_revenue,
			(SELECT COALESCE(SUM(discount_amount),0) FROM orders WHERE status!='cancelled') AS total_discount,
			(SELECT ROUND(AVG(total_price)::numeric, 2) FROM orders WHERE status!='cancelled') AS aov,
			(SELECT COUNT(*) FROM users WHERE role='customer')                          AS customers,
			(SELECT COUNT(DISTINCT user_id) FROM orders)                               AS buyers,
			(SELECT COUNT(*) FROM products WHERE is_active=true)                       AS active_products,
			(SELECT COUNT(*) FROM products WHERE is_on_sale=true AND is_active=true)   AS sale_products
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	if len(rows) == 0 {
		return ReportSheet{}, nil
	}
	r0 := rows[0]
	out := [][]any{
		{"Всего заказов", r0[0]},
		{"Доставлено", r0[1]},
		{"Отменено", r0[2]},
		{"Валовая выручка (₽)", r0[3]},
		{"Сумма скидок (₽)", r0[4]},
		{"Средний чек (AOV, ₽)", r0[5]},
		{"Клиентов всего", r0[6]},
		{"Активных покупателей", r0[7]},
		{"Активных товаров", r0[8]},
		{"На распродаже", r0[9]},
	}
	return ReportSheet{Name: "Сводка", Headers: []string{"Показатель", "Значение"}, Rows: out}, nil
}

func (r *ReportsRepo) RevenueDaily() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT TO_CHAR(o.created_at::date, 'DD.MM.YYYY')         AS day,
		       COUNT(*)                                          AS orders,
		       ROUND(SUM(o.total_price)::numeric, 2)             AS revenue,
		       ROUND(AVG(o.total_price)::numeric, 2)             AS aov
		FROM orders o
		WHERE o.status != 'cancelled'
		  AND o.created_at >= CURRENT_DATE - INTERVAL '90 days'
		GROUP BY o.created_at::date
		ORDER BY o.created_at::date
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "Выручка по дням", Headers: []string{"Дата", "Заказов", "Выручка ₽", "Средний чек ₽"}, Rows: rows}, nil
}

func (r *ReportsRepo) RevenueByCategory() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT c.name,
		       COUNT(DISTINCT o.id)                              AS orders,
		       SUM(oi.quantity)                                  AS units,
		       ROUND(SUM(oi.price_at_order * oi.quantity)::numeric, 2) AS revenue
		FROM order_items oi
		JOIN orders   o  ON o.id  = oi.order_id
		JOIN products p  ON p.id  = oi.product_id
		JOIN categories c ON c.id = p.category_id
		WHERE o.status != 'cancelled'
		GROUP BY c.name
		ORDER BY revenue DESC NULLS LAST
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "Выручка по категориям", Headers: []string{"Категория", "Заказов", "Шт.", "Выручка ₽"}, Rows: rows}, nil
}

func (r *ReportsRepo) RevenueByCity() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT a.city,
		       COUNT(o.id)                                       AS orders,
		       ROUND(SUM(o.total_price)::numeric, 2)             AS revenue,
		       ROUND(AVG(o.total_price)::numeric, 2)             AS aov,
		       COUNT(DISTINCT o.user_id)                          AS customers
		FROM orders o
		JOIN addresses a ON a.id = o.address_id
		WHERE o.status != 'cancelled'
		GROUP BY a.city
		ORDER BY revenue DESC NULLS LAST
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "Выручка по городам", Headers: []string{"Город", "Заказов", "Выручка ₽", "Средний чек ₽", "Клиентов"}, Rows: rows}, nil
}

func (r *ReportsRepo) TopProducts() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT p.name,
		       c.name                                            AS category,
		       SUM(oi.quantity)                                  AS units,
		       ROUND(SUM(oi.price_at_order * oi.quantity)::numeric, 2) AS revenue,
		       ROUND(AVG(oi.price_at_order)::numeric, 2)         AS avg_price
		FROM order_items oi
		JOIN orders   o ON o.id = oi.order_id
		JOIN products p ON p.id = oi.product_id
		JOIN categories c ON c.id = p.category_id
		WHERE o.status != 'cancelled'
		GROUP BY p.id, p.name, c.name
		ORDER BY revenue DESC NULLS LAST
		LIMIT 30
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "Топ-30 товаров", Headers: []string{"Товар", "Категория", "Шт.", "Выручка ₽", "Средняя цена ₽"}, Rows: rows}, nil
}

func (r *ReportsRepo) DeadStock() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT p.name,
		       c.name                       AS category,
		       p.price,
		       p.is_on_sale,
		       p.created_at::date           AS created
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.is_active = true
		  AND NOT EXISTS (
		    SELECT 1 FROM order_items oi
		    JOIN orders o ON o.id = oi.order_id
		    WHERE oi.product_id = p.id
		      AND o.created_at >= CURRENT_DATE - INTERVAL '60 days'
		      AND o.status != 'cancelled'
		  )
		ORDER BY p.created_at ASC
		LIMIT 100
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "Мёртвый сток (60 дней)", Headers: []string{"Товар", "Категория", "Цена ₽", "На распродаже", "Заведён"}, Rows: rows}, nil
}

func (r *ReportsRepo) OrderFunnel() (ReportSheet, error) {
	rows, err := r.queryRows(`
		WITH s AS (
		  SELECT status, COUNT(*) AS cnt FROM orders GROUP BY status
		), tot AS (SELECT SUM(cnt) AS t FROM s)
		SELECT
		  CASE status
		    WHEN 'pending'   THEN 'Ожидает'
		    WHEN 'confirmed' THEN 'Подтверждён'
		    WHEN 'shipped'   THEN 'Отправлен'
		    WHEN 'delivered' THEN 'Доставлен'
		    WHEN 'cancelled' THEN 'Отменён'
		    ELSE status
		  END                                                AS status,
		  cnt                                                AS count,
		  ROUND(cnt::numeric / tot.t * 100, 2)               AS percent
		FROM s, tot
		ORDER BY count DESC
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "Воронка статусов", Headers: []string{"Статус", "Кол-во", "Доля %"}, Rows: rows}, nil
}

func (r *ReportsRepo) CancellationByCategory() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT c.name,
		       COUNT(*) FILTER (WHERE o.status='cancelled') AS cancelled,
		       COUNT(*)                                      AS total,
		       ROUND(
		         COUNT(*) FILTER (WHERE o.status='cancelled')::numeric
		         / NULLIF(COUNT(*),0) * 100, 2)              AS cancel_rate
		FROM orders o
		JOIN order_items oi ON oi.order_id = o.id
		JOIN products p     ON p.id = oi.product_id
		JOIN categories c   ON c.id = p.category_id
		GROUP BY c.name
		ORDER BY cancel_rate DESC NULLS LAST
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "Отмены по категориям", Headers: []string{"Категория", "Отменено", "Всего", "Доля отмен %"}, Rows: rows}, nil
}

func (r *ReportsRepo) PromoROI() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT pc.code,
		       pc.discount_type,
		       pc.discount_value,
		       COUNT(o.id)                                              AS activations,
		       ROUND(COALESCE(SUM(o.total_price),0)::numeric, 2)        AS revenue_generated,
		       ROUND(COALESCE(SUM(o.discount_amount),0)::numeric, 2)    AS discount_given,
		       CASE WHEN COALESCE(SUM(o.discount_amount),0) > 0
		            THEN ROUND(SUM(o.total_price)::numeric / SUM(o.discount_amount), 2)
		            ELSE NULL END                                       AS roi
		FROM promo_codes pc
		LEFT JOIN orders o ON o.promo_code_id = pc.id AND o.status != 'cancelled'
		GROUP BY pc.id, pc.code, pc.discount_type, pc.discount_value
		ORDER BY revenue_generated DESC NULLS LAST
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "ROI промокодов", Headers: []string{"Код", "Тип", "Значение", "Активаций", "Выручка ₽", "Скидка ₽", "ROI ×"}, Rows: rows}, nil
}

func (r *ReportsRepo) TopCustomers() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT u.email,
		       COALESCE(u.name,'')                              AS name,
		       COUNT(o.id)                                      AS orders,
		       ROUND(SUM(o.total_price)::numeric, 2)            AS spent,
		       ROUND(AVG(o.total_price)::numeric, 2)            AS aov,
		       MAX(o.created_at)::date                          AS last_order
		FROM users u
		JOIN orders o ON o.user_id = u.id AND o.status != 'cancelled'
		WHERE u.role = 'customer'
		GROUP BY u.id, u.email, u.name
		ORDER BY spent DESC NULLS LAST
		LIMIT 20
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "Топ-20 клиентов", Headers: []string{"Email", "Имя", "Заказов", "Потратил ₽", "Средний чек ₽", "Последний заказ"}, Rows: rows}, nil
}

func (r *ReportsRepo) RFM() (ReportSheet, error) {
	rows, err := r.queryRows(`
		WITH base AS (
		  SELECT u.id, u.email,
		         (CURRENT_DATE - MAX(o.created_at::date)) AS recency_days,
		         COUNT(o.id)                              AS frequency,
		         SUM(o.total_price)                       AS monetary
		  FROM users u
		  JOIN orders o ON o.user_id = u.id AND o.status != 'cancelled'
		  WHERE u.role = 'customer'
		  GROUP BY u.id, u.email
		)
		SELECT email,
		       recency_days,
		       frequency,
		       ROUND(monetary::numeric, 2)                AS monetary,
		       CASE
		         WHEN recency_days <= 14 AND frequency >= 3 THEN 'VIP'
		         WHEN recency_days <= 30 AND frequency >= 2 THEN 'Лояльный'
		         WHEN recency_days <= 30                    THEN 'Новый'
		         WHEN recency_days <= 60                    THEN 'Уходящий'
		         ELSE                                            'Потерянный'
		       END                                        AS segment
		FROM base
		ORDER BY monetary DESC NULLS LAST
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "RFM-сегменты", Headers: []string{"Email", "Дней с последнего", "Заказов", "Сумма ₽", "Сегмент"}, Rows: rows}, nil
}

func (r *ReportsRepo) Geo() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT a.city,
		       COUNT(DISTINCT u.id)        AS customers,
		       COUNT(DISTINCT o.id)        AS orders,
		       ROUND(COALESCE(SUM(o.total_price),0)::numeric, 2) AS revenue
		FROM addresses a
		JOIN users u   ON u.id = a.user_id AND u.role = 'customer'
		LEFT JOIN orders o ON o.address_id = a.id AND o.status != 'cancelled'
		GROUP BY a.city
		ORDER BY revenue DESC NULLS LAST
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "География", Headers: []string{"Город", "Клиентов", "Заказов", "Выручка ₽"}, Rows: rows}, nil
}

func (r *ReportsRepo) TopWishlisted() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT p.name,
		       c.name                                            AS category,
		       COUNT(w.user_id)                                  AS wishes,
		       COALESCE(s.units, 0)                              AS units_sold,
		       CASE WHEN COUNT(w.user_id) > 0
		            THEN ROUND(COALESCE(s.units,0)::numeric / COUNT(w.user_id) * 100, 2)
		            ELSE 0 END                                   AS conversion_pct
		FROM wishlist w
		JOIN products p ON p.id = w.product_id
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN LATERAL (
		  SELECT SUM(oi.quantity)::int AS units
		  FROM order_items oi JOIN orders o ON o.id = oi.order_id
		  WHERE oi.product_id = p.id AND o.status != 'cancelled'
		) s ON true
		GROUP BY p.id, p.name, c.name, s.units
		ORDER BY wishes DESC NULLS LAST
		LIMIT 30
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "Топ wishlist", Headers: []string{"Товар", "Категория", "В избранном", "Куплено шт.", "Конверсия %"}, Rows: rows}, nil
}

func (r *ReportsRepo) SaleVsRegular() (ReportSheet, error) {
	rows, err := r.queryRows(`
		WITH t AS (
		  SELECT p.is_on_sale,
		         SUM(oi.quantity)                                  AS units,
		         SUM(oi.price_at_order * oi.quantity)              AS revenue,
		         AVG(oi.price_at_order)                            AS avg_price
		  FROM order_items oi
		  JOIN orders   o ON o.id = oi.order_id
		  JOIN products p ON p.id = oi.product_id
		  WHERE o.status != 'cancelled'
		  GROUP BY p.is_on_sale
		)
		SELECT CASE WHEN is_on_sale THEN 'Распродажа' ELSE 'Обычные' END AS kind,
		       units,
		       ROUND(revenue::numeric, 2) AS revenue,
		       ROUND(avg_price::numeric, 2) AS avg_price
		FROM t
		ORDER BY revenue DESC
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{Name: "Sale vs обычные", Headers: []string{"Тип", "Шт.", "Выручка ₽", "Средняя цена ₽"}, Rows: rows}, nil
}

// AllSheets returns the ordered list of every report used by the Excel handler.
func (r *ReportsRepo) AllSheets() ([]ReportSheet, error) {
	loaders := []func() (ReportSheet, error){
		r.Summary,
		r.RevenueDaily,
		r.RevenueByCategory,
		r.RevenueByCity,
		r.TopProducts,
		r.SaleVsRegular,
		r.DeadStock,
		r.OrderFunnel,
		r.CancellationByCategory,
		r.PromoROI,
		r.TopCustomers,
		r.RFM,
		r.Geo,
		r.TopWishlisted,
	}
	out := make([]ReportSheet, 0, len(loaders))
	for _, fn := range loaders {
		s, err := fn()
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}
