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
		`INSERT INTO orders(user_id,address_id,promo_code_id,status,total_price,discount_amount,delivery_method,delivery_cost,recipient_name,payment_method,payment_status)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id`,
		o.UserID, o.AddressID, o.PromoCodeID, o.Status, o.TotalPrice, o.DiscountAmount,
		o.DeliveryMethod, o.DeliveryCost, o.RecipientName, o.PaymentMethod, o.PaymentStatus,
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
		`SELECT id,user_id,address_id,promo_code_id,status,total_price,discount_amount,created_at,delivery_method,delivery_cost,recipient_name,payment_method,payment_status
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
		`SELECT id,user_id,address_id,promo_code_id,status,total_price,discount_amount,created_at,delivery_method,delivery_cost,recipient_name,payment_method,payment_status
		 FROM orders WHERE id=$1`, id,
	).Scan(&o.ID, &o.UserID, &o.AddressID, &promoID, &o.Status, &o.TotalPrice, &o.DiscountAmount, &o.CreatedAt,
		&o.DeliveryMethod, &o.DeliveryCost, &o.RecipientName, &o.PaymentMethod, &o.PaymentStatus)
	if err != nil {
		return nil, err
	}
	if promoID.Valid {
		o.PromoCodeID = &promoID.Int64 // only set the pointer when a promo was used
	}
	// Hydrate each line with product name/type/size and a thumbnail so the order
	// card can render the item without extra requests.
	rows, err := r.db.Query(
		`SELECT oi.id,oi.order_id,oi.product_id,oi.product_size_id,oi.quantity,oi.price_at_order,
		        COALESCE(p.name,''), COALESCE(c.type_name,''), COALESCE(ps.size,''),
		        COALESCE((SELECT image_path FROM product_images WHERE product_id=oi.product_id
		                  ORDER BY is_primary DESC, sort_order LIMIT 1),'')
		 FROM order_items oi
		 LEFT JOIN products p ON p.id=oi.product_id
		 LEFT JOIN categories c ON c.id=p.category_id
		 LEFT JOIN product_sizes ps ON ps.id=oi.product_size_id
		 WHERE oi.order_id=$1`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	o.Items = make([]model.OrderItem, 0)
	for rows.Next() {
		var item model.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductID,
			&item.ProductSizeID, &item.Quantity, &item.PriceAtOrder,
			&item.ProductName, &item.TypeName, &item.Size, &item.ImagePath); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Best-effort hydration of the delivery address for the order card.
	var a model.Address
	if err := r.db.QueryRow(
		`SELECT id,user_id,city,street,house,COALESCE(apartment,''),COALESCE(zip_code,''),is_default
		 FROM addresses WHERE id=$1`, o.AddressID,
	).Scan(&a.ID, &a.UserID, &a.City, &a.Street, &a.House, &a.Apartment, &a.ZipCode, &a.IsDefault); err == nil {
		o.Address = &a
	}
	return o, nil
}

// GetAll is the admin order listing with optional status and date-range filters.
// Filters are stitched in dynamically: the WHERE 1=1 seed lets every optional
// clause be appended with a leading AND, and the $N placeholder counter (i) is
// advanced only for the predicates actually present so positional args line up.
func (r *OrderRepo) GetAll(status, dateFrom, dateTo string) ([]model.Order, error) {
	query := `SELECT id,user_id,address_id,promo_code_id,status,total_price,discount_amount,created_at,delivery_method,delivery_cost,recipient_name,payment_method,payment_status
	          FROM orders WHERE 1=1`
	args := []any{}
	i := 1
	if status != "" {
		query += fmt.Sprintf(" AND status=$%d", i)
		args = append(args, status)
		i++
	}
	// created_at is stored in UTC, but the UI shows (and the admin picks) Moscow dates.
	// Compare on the Moscow calendar date so the filter matches the dates on screen —
	// otherwise an order made late UTC (e.g. 21:43) counts as the previous day here but
	// renders as the next day in the UI, leaking an extra day into the results.
	if dateFrom != "" {
		query += fmt.Sprintf(" AND (created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Europe/Moscow')::date >= $%d::date", i)
		args = append(args, dateFrom)
		i++
	}
	if dateTo != "" {
		query += fmt.Sprintf(" AND (created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Europe/Moscow')::date <= $%d::date", i)
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

// ConfirmPaid marks an order as paid and advances it to confirmed — used by the
// payment-stub confirmation for online card payments.
func (r *OrderRepo) ConfirmPaid(id int64) error {
	_, err := r.db.Exec(`UPDATE orders SET status='confirmed', payment_status='paid' WHERE id=$1`, id)
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
			&o.TotalPrice, &o.DiscountAmount, &o.CreatedAt,
			&o.DeliveryMethod, &o.DeliveryCost, &o.RecipientName, &o.PaymentMethod, &o.PaymentStatus); err != nil {
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
