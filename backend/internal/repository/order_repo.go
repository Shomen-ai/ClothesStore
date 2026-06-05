// Package repository: data-access layer for orders and their line items.
package repository

import (
	"database/sql"
	"fmt"
	"clothes-store/internal/model"
)

// OrderRepo provides CRUD and listing queries over the orders / order_items tables.
type OrderRepo struct{ db *sql.DB }

// NewOrderRepo constructs an OrderRepo bound to the given database handle.
func NewOrderRepo(db *sql.DB) *OrderRepo { return &OrderRepo{db: db} }

// Create persists an order with its line items atomically: it inserts the order
// header, each order_item, decrements the corresponding product_sizes stock, and
// bumps the promo activation counter — all in one transaction so a failure
// anywhere rolls the whole checkout back. The generated IDs are written back into
// the passed structs via RETURNING.
func (r *OrderRepo) Create(o *model.Order, items []model.OrderItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // no-op once Commit succeeds; rolls back on any early return

	// Insert the order header first so its id can be linked to every line item.
	if err = tx.QueryRow(
		`INSERT INTO orders(user_id,address_id,promo_code_id,status,total_price,discount_amount)
		 VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
		o.UserID, o.AddressID, o.PromoCodeID, o.Status, o.TotalPrice, o.DiscountAmount,
	).Scan(&o.ID); err != nil {
		return err
	}
	for i := range items {
		items[i].OrderID = o.ID // back-fill the freshly generated order id
		if err = tx.QueryRow(
			`INSERT INTO order_items(order_id,product_id,product_size_id,quantity,price_at_order)
			 VALUES($1,$2,$3,$4,$5) RETURNING id`,
			items[i].OrderID, items[i].ProductID, items[i].ProductSizeID,
			items[i].Quantity, items[i].PriceAtOrder,
		).Scan(&items[i].ID); err != nil {
			return err
		}
		// Deduct the purchased quantity from the size-specific stock.
		if _, err = tx.Exec(
			`UPDATE product_sizes SET stock_qty = stock_qty - $1 WHERE id = $2`,
			items[i].Quantity, items[i].ProductSizeID,
		); err != nil {
			return err
		}
	}
	// Only count an activation when a promo was actually applied (pointer is nil otherwise).
	if o.PromoCodeID != nil {
		if _, err = tx.Exec(
			`UPDATE promo_codes SET activations_count = activations_count + 1 WHERE id = $1`,
			*o.PromoCodeID,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetByUser returns all orders for one user, newest first. Headers only — line
// items are not loaded here (see GetByID for the full order with items).
func (r *OrderRepo) GetByUser(userID int64) ([]model.Order, error) {
	rows, err := r.db.Query(
		`SELECT id,user_id,address_id,promo_code_id,status,total_price,discount_amount,created_at
		 FROM orders WHERE user_id=$1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

// GetByID loads a single order header plus all of its order_items in a second
// query, returning a fully populated Order.
func (r *OrderRepo) GetByID(id int64) (*model.Order, error) {
	o := &model.Order{}
	var promoID sql.NullInt64 // promo_code_id is nullable; map NULL → nil pointer
	err := r.db.QueryRow(
		`SELECT id,user_id,address_id,promo_code_id,status,total_price,discount_amount,created_at
		 FROM orders WHERE id=$1`, id,
	).Scan(&o.ID, &o.UserID, &o.AddressID, &promoID, &o.Status, &o.TotalPrice, &o.DiscountAmount, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	if promoID.Valid {
		o.PromoCodeID = &promoID.Int64 // only set the pointer when a promo was used
	}
	rows, err := r.db.Query(
		`SELECT id,order_id,product_id,product_size_id,quantity,price_at_order
		 FROM order_items WHERE order_id=$1`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	o.Items = make([]model.OrderItem, 0)
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID,
			&item.ProductSizeID, &item.Quantity, &item.PriceAtOrder); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return o, nil
}

// GetAll is the admin order listing with optional status and date-range filters.
// Filters are stitched in dynamically: the WHERE 1=1 seed lets every optional
// clause be appended with a leading AND, and the $N placeholder counter (i) is
// advanced only for the predicates actually present so positional args line up.
func (r *OrderRepo) GetAll(status, dateFrom, dateTo string) ([]model.Order, error) {
	query := `SELECT id,user_id,address_id,promo_code_id,status,total_price,discount_amount,created_at
	          FROM orders WHERE 1=1`
	args := []any{}
	i := 1
	if status != "" {
		query += fmt.Sprintf(" AND status=$%d", i)
		args = append(args, status)
		i++
	}
	if dateFrom != "" {
		// Inclusive lower bound, compared on the date part only.
		query += fmt.Sprintf(" AND created_at >= $%d::date", i)
		args = append(args, dateFrom)
		i++
	}
	if dateTo != "" {
		// Exclusive upper bound at the start of the next day, so the whole dateTo day is included.
		query += fmt.Sprintf(" AND created_at < ($%d::date + INTERVAL '1 day')", i)
		args = append(args, dateTo)
	}
	query += " ORDER BY created_at DESC"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

// UpdateStatus sets the workflow status (pending/confirmed/shipped/…) of one order.
func (r *OrderRepo) UpdateStatus(id int64, status string) error {
	_, err := r.db.Exec(`UPDATE orders SET status=$1 WHERE id=$2`, status, id)
	return err
}

// scanOrders materialises order header rows, mapping the nullable promo_code_id
// into the optional pointer field. Shared by GetByUser and GetAll.
func scanOrders(rows *sql.Rows) ([]model.Order, error) {
	orders := make([]model.Order, 0)
	for rows.Next() {
		var o model.Order
		var promoID sql.NullInt64
		if err := rows.Scan(&o.ID, &o.UserID, &o.AddressID, &promoID, &o.Status,
			&o.TotalPrice, &o.DiscountAmount, &o.CreatedAt); err != nil {
			return nil, err
		}
		if promoID.Valid {
			o.PromoCodeID = &promoID.Int64
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}
