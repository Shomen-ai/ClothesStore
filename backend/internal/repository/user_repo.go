// Package repository: data-access layer for users and their saved addresses.
package repository

import (
	"database/sql"
	"clothes-store/internal/model"
)

// UserRepo provides queries over the users and addresses tables.
type UserRepo struct{ db *sql.DB }

// NewUserRepo constructs a UserRepo bound to db.
func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

// Create inserts a user and writes the generated id back into u.
func (r *UserRepo) Create(u *model.User) error {
	return r.db.QueryRow(
		`INSERT INTO users(email,password_hash,name,phone,role)
		 VALUES($1,$2,$3,$4,$5) RETURNING id`,
		u.Email, u.PasswordHash, u.Name, u.Phone, u.Role,
	).Scan(&u.ID)
}

// GetByEmail loads a user by email, including the password hash for login
// verification. COALESCE maps a NULL phone to an empty string.
func (r *UserRepo) GetByEmail(email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id,email,password_hash,name,COALESCE(phone,''),role,created_at FROM users WHERE email=$1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetByID loads a user by id (without the password hash, for profile display).
// COALESCE maps a NULL phone to an empty string.
func (r *UserRepo) GetByID(id int64) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id,email,name,COALESCE(phone,''),role,created_at FROM users WHERE id=$1`, id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Phone, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// Update saves editable profile fields (name and phone) for a user.
func (r *UserRepo) Update(u *model.User) error {
	_, err := r.db.Exec(
		`UPDATE users SET name=$1,phone=$2 WHERE id=$3`,
		u.Name, u.Phone, u.ID,
	)
	return err
}

// UpdatePassword replaces the stored password hash for a user.
func (r *UserRepo) UpdatePassword(id int64, hash string) error {
	_, err := r.db.Exec(`UPDATE users SET password_hash=$1 WHERE id=$2`, hash, id)
	return err
}

// CreateAddress inserts an address for a user and writes the generated id back into a.
func (r *UserRepo) CreateAddress(a *model.Address) error {
	return r.db.QueryRow(
		`INSERT INTO addresses(user_id,city,street,house,apartment,zip_code,is_default)
		 VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		a.UserID, a.City, a.Street, a.House, a.Apartment, a.ZipCode, a.IsDefault,
	).Scan(&a.ID)
}

// GetAddresses returns a user's addresses with the default one first
// (ORDER BY is_default DESC puts the true/default ahead of the others).
func (r *UserRepo) GetAddresses(userID int64) ([]model.Address, error) {
	rows, err := r.db.Query(
		`SELECT id,user_id,city,street,house,apartment,zip_code,is_default
		 FROM addresses WHERE user_id=$1 ORDER BY is_default DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	addrs := make([]model.Address, 0)
	for rows.Next() {
		var a model.Address
		if err := rows.Scan(&a.ID, &a.UserID, &a.City, &a.Street, &a.House,
			&a.Apartment, &a.ZipCode, &a.IsDefault); err != nil {
			return nil, err
		}
		addrs = append(addrs, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return addrs, nil
}

// UpdateAddress edits an address; the user_id predicate scopes the update to the
// owner so one user cannot modify another's address.
func (r *UserRepo) UpdateAddress(a *model.Address) error {
	_, err := r.db.Exec(
		`UPDATE addresses SET city=$1,street=$2,house=$3,apartment=$4,zip_code=$5,is_default=$6
		 WHERE id=$7 AND user_id=$8`,
		a.City, a.Street, a.House, a.Apartment, a.ZipCode, a.IsDefault, a.ID, a.UserID,
	)
	return err
}

// DeleteAddress removes an address, scoped to its owner via the user_id predicate.
func (r *UserRepo) DeleteAddress(id, userID int64) error {
	_, err := r.db.Exec(`DELETE FROM addresses WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

// SetDefaultAddress makes one address the user's default. It runs in a
// transaction: first clear the flag on all of the user's addresses, then set it
// on the target, so at most one default ever exists.
func (r *UserRepo) SetDefaultAddress(id, userID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // no-op after Commit; undoes the partial update on error
	if _, err = tx.Exec(`UPDATE addresses SET is_default=false WHERE user_id=$1`, userID); err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE addresses SET is_default=true WHERE id=$1 AND user_id=$2`, id, userID); err != nil {
		return err
	}
	return tx.Commit()
}
