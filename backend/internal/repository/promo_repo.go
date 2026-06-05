// Package repository: data-access layer for promo codes.
package repository

import (
	"database/sql"
	"clothes-store/internal/model"
)

// PromoRepo provides queries over the promo_codes table.
type PromoRepo struct{ db *sql.DB }

// NewPromoRepo constructs a PromoRepo bound to db.
func NewPromoRepo(db *sql.DB) *PromoRepo { return &PromoRepo{db: db} }

// GetByCode looks up a promo code by its code string. max_activations and
// expires_at are nullable (unlimited / never expires), so they are scanned into
// sql.Null* and mapped to the model's optional pointer fields only when present.
func (r *PromoRepo) GetByCode(code string) (*model.PromoCode, error) {
	p := &model.PromoCode{}
	var maxAct sql.NullInt64
	var expiresAt sql.NullTime
	err := r.db.QueryRow(
		`SELECT id,code,discount_type,discount_value,max_activations,activations_count,expires_at,is_active,created_at
		 FROM promo_codes WHERE code=$1`, code,
	).Scan(&p.ID, &p.Code, &p.DiscountType, &p.DiscountValue, &maxAct, &p.ActivationsCount,
		&expiresAt, &p.IsActive, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	if maxAct.Valid {
		v := int(maxAct.Int64) // narrow int64 → int before taking its address
		p.MaxActivations = &v
	}
	if expiresAt.Valid {
		p.ExpiresAt = &expiresAt.Time
	}
	return p, nil
}

// GetAll returns every promo code, newest first, applying the same nullable →
// pointer mapping as GetByCode for max_activations and expires_at.
func (r *PromoRepo) GetAll() ([]model.PromoCode, error) {
	rows, err := r.db.Query(
		`SELECT id,code,discount_type,discount_value,max_activations,activations_count,expires_at,is_active,created_at
		 FROM promo_codes ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	promos := make([]model.PromoCode, 0)
	for rows.Next() {
		var p model.PromoCode
		var maxAct sql.NullInt64
		var expiresAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.Code, &p.DiscountType, &p.DiscountValue, &maxAct, &p.ActivationsCount,
			&expiresAt, &p.IsActive, &p.CreatedAt); err != nil {
			return nil, err
		}
		if maxAct.Valid {
			v := int(maxAct.Int64) // narrow int64 → int before taking its address
			p.MaxActivations = &v
		}
		if expiresAt.Valid {
			p.ExpiresAt = &expiresAt.Time
		}
		promos = append(promos, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return promos, nil
}

// Create inserts a promo code and writes the generated id back into p. Nil
// MaxActivations / ExpiresAt pointers are bound as SQL NULL (unlimited / no expiry).
func (r *PromoRepo) Create(p *model.PromoCode) error {
	return r.db.QueryRow(
		`INSERT INTO promo_codes(code,discount_type,discount_value,max_activations,expires_at,is_active)
		 VALUES($1,$2,$3,$4,$5,$6) RETURNING id`,
		p.Code, p.DiscountType, p.DiscountValue, p.MaxActivations, p.ExpiresAt, p.IsActive,
	).Scan(&p.ID)
}

// Deactivate soft-disables a promo code (sets is_active=false) without deleting it.
func (r *PromoRepo) Deactivate(id int64) error {
	_, err := r.db.Exec(`UPDATE promo_codes SET is_active=false WHERE id=$1`, id)
	return err
}

// Delete permanently removes a promo code by id.
func (r *PromoRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM promo_codes WHERE id=$1`, id)
	return err
}

// IncrementActivations atomically bumps the usage counter by one.
func (r *PromoRepo) IncrementActivations(id int64) error {
	_, err := r.db.Exec(`UPDATE promo_codes SET activations_count=activations_count+1 WHERE id=$1`, id)
	return err
}
