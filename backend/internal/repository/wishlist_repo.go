// Package repository: data-access layer for the per-user wishlist.
package repository

import (
	"database/sql"
	"clothes-store/internal/model"
)

// WishlistRepo provides queries over the wishlist join table (user ↔ product).
type WishlistRepo struct{ db *sql.DB }

// NewWishlistRepo constructs a WishlistRepo bound to db.
func NewWishlistRepo(db *sql.DB) *WishlistRepo { return &WishlistRepo{db: db} }

// Get returns the products on a user's wishlist, most recently added first,
// joining wishlist rows to their product records.
func (r *WishlistRepo) Get(userID int64) ([]model.Product, error) {
	rows, err := r.db.Query(
		`SELECT p.id,p.category_id,p.name,p.description,p.price,p.is_active,p.created_at
		 FROM wishlist w JOIN products p ON p.id=w.product_id
		 WHERE w.user_id=$1 ORDER BY w.created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	products := make([]model.Product, 0)
	for rows.Next() {
		var p model.Product
		if err := rows.Scan(&p.ID, &p.CategoryID, &p.Name, &p.Description,
			&p.Price, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}

// Add inserts a (user, product) wishlist entry.
func (r *WishlistRepo) Add(userID, productID int64) error {
	_, err := r.db.Exec(
		`INSERT INTO wishlist(user_id,product_id) VALUES($1,$2)`, userID, productID,
	)
	return err
}

// Remove deletes a product from the user's wishlist.
func (r *WishlistRepo) Remove(userID, productID int64) error {
	_, err := r.db.Exec(`DELETE FROM wishlist WHERE user_id=$1 AND product_id=$2`, userID, productID)
	return err
}
