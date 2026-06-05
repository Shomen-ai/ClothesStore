// Package repository: data-access layer for product reviews (customer + admin).
package repository

import (
	"database/sql"
	"errors"

	"clothes-store/internal/model"
)

// ReviewRepo provides queries over the product_reviews table.
type ReviewRepo struct{ db *sql.DB }

// NewReviewRepo constructs a ReviewRepo bound to db.
func NewReviewRepo(db *sql.DB) *ReviewRepo { return &ReviewRepo{db: db} }

// ListForProduct returns visible reviews (newest first), joining the display name.
func (r *ReviewRepo) ListForProduct(productID int64) ([]model.Review, error) {
	// Display name falls back to email when the profile name is blank:
	// NULLIF('', '') turns an empty name into NULL so COALESCE picks the email.
	rows, err := r.db.Query(`
		SELECT r.id, r.user_id, r.product_id, r.rating, r.text, r.created_at, r.updated_at,
		       COALESCE(NULLIF(u.name, ''), u.email) AS user_name
		FROM product_reviews r
		JOIN users u ON u.id = r.user_id
		WHERE r.product_id = $1 AND r.is_hidden = FALSE
		ORDER BY r.created_at DESC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.Review, 0)
	for rows.Next() {
		var rv model.Review
		if err := rows.Scan(&rv.ID, &rv.UserID, &rv.ProductID, &rv.Rating, &rv.Text,
			&rv.CreatedAt, &rv.UpdatedAt, &rv.UserName); err != nil {
			return nil, err
		}
		out = append(out, rv)
	}
	return out, rows.Err()
}

// SummaryForProduct returns avg rating (rounded to 1 decimal) and visible count.
func (r *ReviewRepo) SummaryForProduct(productID int64) (model.ReviewSummary, error) {
	var s model.ReviewSummary
	// COALESCE(..., 0) yields 0.0 for a product with no visible reviews; the
	// ::float8 cast matches the Go float64 destination.
	err := r.db.QueryRow(`
		SELECT COALESCE(ROUND(AVG(rating)::numeric, 1), 0)::float8, COUNT(*)
		FROM product_reviews
		WHERE product_id = $1 AND is_hidden = FALSE`, productID).Scan(&s.AvgRating, &s.Count)
	return s, err
}

// IsEligible reports whether the user has a delivered order containing the product.
func (r *ReviewRepo) IsEligible(userID, productID int64) (bool, error) {
	var ok bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
		  SELECT 1 FROM order_items oi
		  JOIN orders o ON o.id = oi.order_id
		  WHERE o.user_id = $1 AND oi.product_id = $2 AND o.status = 'delivered'
		)`, userID, productID).Scan(&ok)
	return ok, err
}

// GetMine returns the user's own review for a product (or nil if none).
func (r *ReviewRepo) GetMine(userID, productID int64) (*model.Review, error) {
	var rv model.Review
	err := r.db.QueryRow(`
		SELECT id, user_id, product_id, rating, text, created_at, updated_at
		FROM product_reviews
		WHERE user_id = $1 AND product_id = $2`, userID, productID).
		Scan(&rv.ID, &rv.UserID, &rv.ProductID, &rv.Rating, &rv.Text, &rv.CreatedAt, &rv.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// Create inserts a new review. The UNIQUE(user_id, product_id) constraint returns
// the standard Postgres unique-violation error if the user already left one.
func (r *ReviewRepo) Create(userID, productID int64, rating int, text string) (*model.Review, error) {
	var rv model.Review
	rv.UserID = userID
	rv.ProductID = productID
	rv.Rating = rating
	rv.Text = text
	err := r.db.QueryRow(`
		INSERT INTO product_reviews(user_id, product_id, rating, text)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`,
		userID, productID, rating, text).
		Scan(&rv.ID, &rv.CreatedAt, &rv.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// Update modifies own review; returns sql.ErrNoRows if nothing matched (not owner or missing).
func (r *ReviewRepo) Update(reviewID, userID int64, rating int, text string) (*model.Review, error) {
	var rv model.Review
	err := r.db.QueryRow(`
		UPDATE product_reviews
		SET rating = $1, text = $2, updated_at = NOW()
		WHERE id = $3 AND user_id = $4
		RETURNING id, user_id, product_id, rating, text, created_at, updated_at`,
		rating, text, reviewID, userID).
		Scan(&rv.ID, &rv.UserID, &rv.ProductID, &rv.Rating, &rv.Text, &rv.CreatedAt, &rv.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rv, nil
}

// Delete removes own review.
func (r *ReviewRepo) Delete(reviewID, userID int64) error {
	res, err := r.db.Exec(`DELETE FROM product_reviews WHERE id = $1 AND user_id = $2`, reviewID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ---- admin ----

// AdminList lists all reviews (optionally filtered by visibility) with user/product joins.
func (r *ReviewRepo) AdminList(hidden *bool) ([]model.Review, error) {
	q := `SELECT r.id, r.user_id, r.product_id, r.rating, r.text, r.is_hidden,
	             r.created_at, r.updated_at,
	             COALESCE(NULLIF(u.name, ''), u.email) AS user_name,
	             u.email, p.name AS product_name
	      FROM product_reviews r
	      JOIN users u    ON u.id = r.user_id
	      JOIN products p ON p.id = r.product_id`
	args := []any{}
	if hidden != nil {
		q += ` WHERE r.is_hidden = $1`
		args = append(args, *hidden)
	}
	q += ` ORDER BY r.created_at DESC`

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.Review, 0)
	for rows.Next() {
		var rv model.Review
		if err := rows.Scan(&rv.ID, &rv.UserID, &rv.ProductID, &rv.Rating, &rv.Text,
			&rv.IsHidden, &rv.CreatedAt, &rv.UpdatedAt,
			&rv.UserName, &rv.UserEmail, &rv.ProductName); err != nil {
			return nil, err
		}
		out = append(out, rv)
	}
	return out, rows.Err()
}

// AdminSetHidden toggles a review's visibility. Returns sql.ErrNoRows when no
// review matched the id (zero rows affected), so callers can surface a 404.
func (r *ReviewRepo) AdminSetHidden(reviewID int64, hidden bool) error {
	res, err := r.db.Exec(`UPDATE product_reviews SET is_hidden = $1, updated_at = NOW() WHERE id = $2`, hidden, reviewID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// AdminDelete removes any review by id (no owner check). Returns sql.ErrNoRows
// when the id did not exist.
func (r *ReviewRepo) AdminDelete(reviewID int64) error {
	res, err := r.db.Exec(`DELETE FROM product_reviews WHERE id = $1`, reviewID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
