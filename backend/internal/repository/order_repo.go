package repository

import (
	"database/sql"
	"fmt"
	"clothes-store/internal/model"
)

type OrderRepo struct{ db *sql.DB }

func NewOrderRepo(db *sql.DB) *OrderRepo { return &OrderRepo{db: db} }

func (r *OrderRepo) Create(o *model.Order, items []model.OrderItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err = tx.QueryRow(
		`INSERT INTO orders(user_id,address_id,promo_code_id,status,total_price,discount_amount)
		 VALUES(:1,:2,:3,:4,:5,:6) RETURNING id INTO :7`,
		o.UserID, o.AddressID, o.PromoCodeID, o.Status, o.TotalPrice, o.DiscountAmount, &o.ID,
	).Err(); err != nil {
		return err
	}
	for i := range items {
		items[i].OrderID = o.ID
		if err = tx.QueryRow(
			`INSERT INTO order_items(order_id,product_id,product_size_id,quantity,price_at_order)
			 VALUES(:1,:2,:3,:4,:5) RETURNING id INTO :6`,
			items[i].OrderID, items[i].ProductID, items[i].ProductSizeID,
			items[i].Quantity, items[i].PriceAtOrder, &items[i].ID,
		).Err(); err != nil {
			return err
		}
		if _, err = tx.Exec(
			`UPDATE product_sizes SET stock_qty = stock_qty - :1 WHERE id = :2`,
			items[i].Quantity, items[i].ProductSizeID,
		); err != nil {
			return err
		}
	}
	if o.PromoCodeID != nil {
		if _, err = tx.Exec(
			`UPDATE promo_codes SET activations_count = activations_count + 1 WHERE id = :1`,
			*o.PromoCodeID,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *OrderRepo) GetByUser(userID int64) ([]model.Order, error) {
	rows, err := r.db.Query(
		`SELECT id,user_id,address_id,promo_code_id,status,total_price,discount_amount,created_at
		 FROM orders WHERE user_id=:1 ORDER BY created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanOrders(rows)
}

func (r *OrderRepo) GetByID(id int64) (*model.Order, error) {
	o := &model.Order{}
	var promoID sql.NullInt64
	err := r.db.QueryRow(
		`SELECT id,user_id,address_id,promo_code_id,status,total_price,discount_amount,created_at
		 FROM orders WHERE id=:1`, id,
	).Scan(&o.ID, &o.UserID, &o.AddressID, &promoID, &o.Status, &o.TotalPrice, &o.DiscountAmount, &o.CreatedAt)
	if err != nil {
		return nil, err
	}
	if promoID.Valid {
		o.PromoCodeID = &promoID.Int64
	}
	rows, err := r.db.Query(
		`SELECT id,order_id,product_id,product_size_id,quantity,price_at_order
		 FROM order_items WHERE order_id=:1`, id,
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

func (r *OrderRepo) GetAll(status, dateFrom, dateTo string) ([]model.Order, error) {
	query := `SELECT id,user_id,address_id,promo_code_id,status,total_price,discount_amount,created_at
	          FROM orders WHERE 1=1`
	args := []any{}
	i := 1
	if status != "" {
		query += fmt.Sprintf(" AND status=:%d", i)
		args = append(args, status)
		i++
	}
	if dateFrom != "" {
		query += fmt.Sprintf(" AND created_at >= TO_TIMESTAMP(:%d,'YYYY-MM-DD')", i)
		args = append(args, dateFrom)
		i++
	}
	if dateTo != "" {
		query += fmt.Sprintf(" AND created_at <= TO_TIMESTAMP(:%d,'YYYY-MM-DD')+1", i)
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

func (r *OrderRepo) UpdateStatus(id int64, status string) error {
	_, err := r.db.Exec(`UPDATE orders SET status=:1 WHERE id=:2`, status, id)
	return err
}

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
