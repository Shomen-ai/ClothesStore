package repository

import (
	"database/sql"
)

// ReportsRepo aggregates all read-only queries used by the Excel report.
// Each method returns a ReportSheet whose ColTypes drive number formatting
// and alignment in the handler, so the same struct shape works for tables
// of money, percentages, counts and dates without per-report ceremony.
type ReportsRepo struct{ db *sql.DB }

func NewReportsRepo(db *sql.DB) *ReportsRepo { return &ReportsRepo{db: db} }

// ColType tells the handler how a value should be formatted in Excel.
type ColType int

const (
	ColText ColType = iota
	ColInt
	ColMoney   // numeric, "# ##0 ₽"
	ColPercent // value already in percent units, "0.0\"%\""
	ColDate    // already-formatted date string; right-aligned, mono
)

// ReportSheet is one rendered Excel sheet: a titled table whose columns are
// described by parallel Headers/ColTypes slices, with optional totalled columns.
type ReportSheet struct {
	Name       string
	Headers    []string
	ColTypes   []ColType
	Rows       [][]any
	TotalsCols []int // 0-indexed columns to SUM in a final "Итого" row
}

// queryRows runs a query and returns rows as [][]any, normalising []byte to string.
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
		// lib/pq returns numeric/text columns as []byte; convert to string so
		// the Excel layer sees readable values rather than byte slices.
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

// Summary is a flat key/value list; rendering as two text columns keeps it
// human-readable even with mixed numeric types in the value column.
func (r *ReportsRepo) Summary() (ReportSheet, error) {
	// Every sub-aggregate is clamped at created_at <= NOW() so seeded future
	// orders never inflate the totals; gross_revenue / discount / AOV also drop
	// cancelled orders. COALESCE keeps sums at 0 when no rows match.
	rows, err := r.queryRows(`
		SELECT
			(SELECT COUNT(*) FROM orders WHERE created_at <= NOW())                                            AS total_orders,
			(SELECT COUNT(*) FROM orders WHERE status='delivered' AND created_at <= NOW())                     AS delivered,
			(SELECT COUNT(*) FROM orders WHERE status='cancelled' AND created_at <= NOW())                     AS cancelled,
			(SELECT COALESCE(SUM(total_price),0)     FROM orders WHERE status!='cancelled' AND created_at <= NOW()) AS gross_revenue,
			(SELECT COALESCE(SUM(discount_amount),0) FROM orders WHERE status!='cancelled' AND created_at <= NOW()) AS total_discount,
			(SELECT ROUND(AVG(total_price)::numeric, 2) FROM orders WHERE status!='cancelled' AND created_at <= NOW()) AS aov,
			(SELECT COUNT(*) FROM users WHERE role='customer')                          AS customers,
			(SELECT COUNT(DISTINCT user_id) FROM orders WHERE created_at <= NOW())     AS buyers,
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
	return ReportSheet{
		Name:     "Сводка",
		Headers:  []string{"Показатель", "Значение"},
		ColTypes: []ColType{ColText, ColText},
		Rows:     out,
	}, nil
}

// RevenueDaily reports orders/revenue/AOV per day over the trailing 90 days,
// excluding cancelled orders and clamping the window at NOW().
func (r *ReportsRepo) RevenueDaily() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT TO_CHAR(o.created_at::date, 'DD.MM.YYYY')         AS day,
		       COUNT(*)                                          AS orders,
		       ROUND(SUM(o.total_price)::numeric, 2)             AS revenue,
		       ROUND(AVG(o.total_price)::numeric, 2)             AS aov
		FROM orders o
		WHERE o.status != 'cancelled'
		  AND o.created_at >= CURRENT_DATE - INTERVAL '90 days'
		  AND o.created_at <= NOW()
		GROUP BY o.created_at::date
		ORDER BY o.created_at::date
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{
		Name:       "Выручка по дням",
		Headers:    []string{"Дата", "Заказов", "Выручка ₽", "Средний чек ₽"},
		ColTypes:   []ColType{ColDate, ColInt, ColMoney, ColMoney},
		Rows:       rows,
		TotalsCols: []int{1, 2},
	}, nil
}

// RevenueByCategory aggregates line items up to their product's category.
// Revenue uses price_at_order * quantity (the price captured at purchase time,
// not the current product price); COUNT(DISTINCT o.id) avoids double-counting
// an order that has several items in the same category.
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
		  AND o.created_at <= NOW()
		GROUP BY c.name
		ORDER BY revenue DESC NULLS LAST
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{
		Name:       "Выручка по категориям",
		Headers:    []string{"Категория", "Заказов", "Шт.", "Выручка ₽"},
		ColTypes:   []ColType{ColText, ColInt, ColInt, ColMoney},
		Rows:       rows,
		TotalsCols: []int{1, 2, 3},
	}, nil
}

// RevenueByCity joins each order to its shipping address and aggregates
// orders/revenue/AOV/unique-customers per city (cancelled excluded, clamped at NOW()).
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
		  AND o.created_at <= NOW()
		GROUP BY a.city
		ORDER BY revenue DESC NULLS LAST
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{
		Name:       "Выручка по городам",
		Headers:    []string{"Город", "Заказов", "Выручка ₽", "Средний чек ₽", "Клиентов"},
		ColTypes:   []ColType{ColText, ColInt, ColMoney, ColMoney, ColInt},
		Rows:       rows,
		TotalsCols: []int{1, 2, 4},
	}, nil
}

// TopProducts ranks the 30 highest-revenue products (revenue = sum of
// price_at_order * quantity), with units sold and average sale price.
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
		  AND o.created_at <= NOW()
		GROUP BY p.id, p.name, c.name
		ORDER BY revenue DESC NULLS LAST
		LIMIT 30
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{
		Name:       "Топ-30 товаров",
		Headers:    []string{"Товар", "Категория", "Шт.", "Выручка ₽", "Средняя цена ₽"},
		ColTypes:   []ColType{ColText, ColText, ColInt, ColMoney, ColMoney},
		Rows:       rows,
		TotalsCols: []int{2, 3},
	}, nil
}

// DeadStock lists up to 100 active products with no non-cancelled sale in the
// last 60 days (NOT EXISTS over recent order_items), oldest first — the slow movers.
func (r *ReportsRepo) DeadStock() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT p.name,
		       c.name                       AS category,
		       p.price,
		       p.is_on_sale,
		       TO_CHAR(p.created_at::date, 'DD.MM.YYYY') AS created
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE p.is_active = true
		  AND NOT EXISTS (
		    SELECT 1 FROM order_items oi
		    JOIN orders o ON o.id = oi.order_id
		    WHERE oi.product_id = p.id
		      AND o.created_at >= CURRENT_DATE - INTERVAL '60 days'
		      AND o.created_at <= NOW()
		      AND o.status != 'cancelled'
		  )
		ORDER BY p.created_at ASC
		LIMIT 100
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	// Normalise the boolean column for nicer text rendering.
	for _, row := range rows {
		if v, ok := row[3].(bool); ok {
			if v {
				row[3] = "да"
			} else {
				row[3] = "нет"
			}
		}
	}
	return ReportSheet{
		Name:     "Мёртвый сток (60 дней)",
		Headers:  []string{"Товар", "Категория", "Цена ₽", "На распродаже", "Заведён"},
		ColTypes: []ColType{ColText, ColText, ColMoney, ColText, ColDate},
		Rows:     rows,
	}, nil
}

// OrderFunnel breaks orders down by status with each status's share of the
// total. The s CTE counts orders per status; tot holds the grand total so the
// cross join (FROM s, tot) can compute cnt/total*100 per row. Status codes are
// translated to Russian labels via CASE.
func (r *ReportsRepo) OrderFunnel() (ReportSheet, error) {
	rows, err := r.queryRows(`
		WITH s AS (
		  SELECT status, COUNT(*) AS cnt FROM orders WHERE created_at <= NOW() GROUP BY status
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
	return ReportSheet{
		Name:       "Воронка статусов",
		Headers:    []string{"Статус", "Кол-во", "Доля %"},
		ColTypes:   []ColType{ColText, ColInt, ColPercent},
		Rows:       rows,
		TotalsCols: []int{1},
	}, nil
}

// CancellationByCategory computes per-category cancel rate. COUNT(*) FILTER
// counts only cancelled rows; NULLIF(COUNT(*),0) guards the division so a
// category with zero rows yields NULL rather than a divide-by-zero error.
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
		WHERE o.created_at <= NOW()
		GROUP BY c.name
		ORDER BY cancel_rate DESC NULLS LAST
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{
		Name:       "Отмены по категориям",
		Headers:    []string{"Категория", "Отменено", "Всего", "Доля отмен %"},
		ColTypes:   []ColType{ColText, ColInt, ColInt, ColPercent},
		Rows:       rows,
		TotalsCols: []int{1, 2},
	}, nil
}

// PromoROI reports, per promo code, the revenue it drove vs the discount it
// cost, plus an ROI ratio (revenue / discount). The LEFT JOIN keeps codes with
// zero redemptions visible; its non-cancelled / NOW() predicates live in the ON
// clause so unredeemed codes still appear (a WHERE would filter them out). ROI
// is NULL when no discount was actually given (avoids divide-by-zero).
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
		LEFT JOIN orders o ON o.promo_code_id = pc.id AND o.status != 'cancelled' AND o.created_at <= NOW()
		GROUP BY pc.id, pc.code, pc.discount_type, pc.discount_value
		ORDER BY revenue_generated DESC NULLS LAST
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{
		Name:       "ROI промокодов",
		Headers:    []string{"Код", "Тип", "Значение", "Активаций", "Выручка ₽", "Скидка ₽", "ROI ×"},
		ColTypes:   []ColType{ColText, ColText, ColText, ColInt, ColMoney, ColMoney, ColText},
		Rows:       rows,
		TotalsCols: []int{3, 4, 5},
	}, nil
}

// TopCustomers ranks the 20 highest-spending customers (sum of non-cancelled
// order totals), with order count, AOV and last-order date.
func (r *ReportsRepo) TopCustomers() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT u.email,
		       COALESCE(u.name,'')                              AS name,
		       COUNT(o.id)                                      AS orders,
		       ROUND(SUM(o.total_price)::numeric, 2)            AS spent,
		       ROUND(AVG(o.total_price)::numeric, 2)            AS aov,
		       TO_CHAR(MAX(o.created_at)::date, 'DD.MM.YYYY')   AS last_order
		FROM users u
		JOIN orders o ON o.user_id = u.id AND o.status != 'cancelled' AND o.created_at <= NOW()
		WHERE u.role = 'customer'
		GROUP BY u.id, u.email, u.name
		ORDER BY spent DESC NULLS LAST
		LIMIT 20
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{
		Name:       "Топ-20 клиентов",
		Headers:    []string{"Email", "Имя", "Заказов", "Потратил ₽", "Средний чек ₽", "Последний заказ"},
		ColTypes:   []ColType{ColText, ColText, ColInt, ColMoney, ColMoney, ColDate},
		Rows:       rows,
		TotalsCols: []int{2, 3},
	}, nil
}

// RFM buckets customers into marketing segments from Recency (days since last
// order), Frequency (order count) and Monetary (total spend). The base CTE
// derives the three metrics per customer; the outer CASE ladder maps them to a
// segment label (VIP / Лояльный / Новый / Уходящий / Потерянный).
func (r *ReportsRepo) RFM() (ReportSheet, error) {
	rows, err := r.queryRows(`
		WITH base AS (
		  SELECT u.id, u.email,
		         (CURRENT_DATE - MAX(o.created_at::date)) AS recency_days,
		         COUNT(o.id)                              AS frequency,
		         SUM(o.total_price)                       AS monetary
		  FROM users u
		  JOIN orders o ON o.user_id = u.id AND o.status != 'cancelled' AND o.created_at <= NOW()
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
	return ReportSheet{
		Name:       "RFM-сегменты",
		Headers:    []string{"Email", "Дней с последнего", "Заказов", "Сумма ₽", "Сегмент"},
		ColTypes:   []ColType{ColText, ColInt, ColInt, ColMoney, ColText},
		Rows:       rows,
		TotalsCols: []int{3},
	}, nil
}

// Geo aggregates customers/orders/revenue per city. The LEFT JOIN to orders
// keeps cities that have registered customers but no orders yet; its filters
// sit in the ON clause so those zero-order cities are not dropped.
func (r *ReportsRepo) Geo() (ReportSheet, error) {
	rows, err := r.queryRows(`
		SELECT a.city,
		       COUNT(DISTINCT u.id)        AS customers,
		       COUNT(DISTINCT o.id)        AS orders,
		       ROUND(COALESCE(SUM(o.total_price),0)::numeric, 2) AS revenue
		FROM addresses a
		JOIN users u   ON u.id = a.user_id AND u.role = 'customer'
		LEFT JOIN orders o ON o.address_id = a.id AND o.status != 'cancelled' AND o.created_at <= NOW()
		GROUP BY a.city
		ORDER BY revenue DESC NULLS LAST
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{
		Name:       "География",
		Headers:    []string{"Город", "Клиентов", "Заказов", "Выручка ₽"},
		ColTypes:   []ColType{ColText, ColInt, ColInt, ColMoney},
		Rows:       rows,
		TotalsCols: []int{1, 2, 3},
	}, nil
}

// TopWishlisted ranks the 30 most-wishlisted products and a wish→purchase
// conversion (units sold / wish count * 100). Units sold come from a LEFT JOIN
// LATERAL subquery (correlated per product) so the per-product sum is computed
// once and reused; conversion guards against zero wishes.
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
		  WHERE oi.product_id = p.id AND o.status != 'cancelled' AND o.created_at <= NOW()
		) s ON true
		GROUP BY p.id, p.name, c.name, s.units
		ORDER BY wishes DESC NULLS LAST
		LIMIT 30
	`)
	if err != nil {
		return ReportSheet{}, err
	}
	return ReportSheet{
		Name:       "Топ wishlist",
		Headers:    []string{"Товар", "Категория", "В избранном", "Куплено шт.", "Конверсия %"},
		ColTypes:   []ColType{ColText, ColText, ColInt, ColInt, ColPercent},
		Rows:       rows,
		TotalsCols: []int{2, 3},
	}, nil
}

// SaleVsRegular compares units/revenue/avg-price of on-sale vs regular-priced
// products, grouping line items by the product's is_on_sale flag.
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
		    AND o.created_at <= NOW()
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
	return ReportSheet{
		Name:       "Sale vs обычные",
		Headers:    []string{"Тип", "Шт.", "Выручка ₽", "Средняя цена ₽"},
		ColTypes:   []ColType{ColText, ColInt, ColMoney, ColMoney},
		Rows:       rows,
		TotalsCols: []int{1, 2},
	}, nil
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

// SummaryKPIs is a flat snapshot used by the Excel "Содержание" cover sheet.
// Keeping it separate avoids forcing the Summary sheet itself into a mixed-type table.
type SummaryKPIs struct {
	TotalRevenue   float64
	TotalOrders    int
	AOV            float64
	Customers      int
	ActiveProducts int
	SaleShare      float64 // percent of active products on sale
}

// KPIs returns the headline metrics for the cover sheet in one round trip:
// total revenue, order count, AOV, customer/active-product counts and the
// share of active products on sale. Each sub-select is COALESCEd to 0 (and
// cast to float8 for the money/ratio fields); SaleShare uses NULLIF to avoid
// dividing by zero when there are no active products.
func (r *ReportsRepo) KPIs() (SummaryKPIs, error) {
	var k SummaryKPIs
	err := r.db.QueryRow(`
		SELECT
		  COALESCE((SELECT SUM(total_price)             FROM orders WHERE status!='cancelled' AND created_at <= NOW()), 0)::float8,
		  COALESCE((SELECT COUNT(*)                     FROM orders WHERE created_at <= NOW()), 0),
		  COALESCE((SELECT AVG(total_price)::numeric    FROM orders WHERE status!='cancelled' AND created_at <= NOW()), 0)::float8,
		  COALESCE((SELECT COUNT(*) FROM users WHERE role='customer'), 0),
		  COALESCE((SELECT COUNT(*) FROM products WHERE is_active=true), 0),
		  COALESCE((
		    SELECT ROUND(
		      100.0 * COUNT(*) FILTER (WHERE is_on_sale=true)
		      / NULLIF(COUNT(*), 0), 1
		    )
		    FROM products WHERE is_active=true
		  ), 0)::float8
	`).Scan(&k.TotalRevenue, &k.TotalOrders, &k.AOV, &k.Customers, &k.ActiveProducts, &k.SaleShare)
	return k, err
}
