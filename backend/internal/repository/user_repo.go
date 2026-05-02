package repository

import (
	"database/sql"
	"clothes-store/internal/model"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(u *model.User) error {
	return r.db.QueryRow(
		`INSERT INTO users(email,password_hash,name,phone,role)
		 VALUES(:1,:2,:3,:4,:5) RETURNING id INTO :6`,
		u.Email, u.PasswordHash, u.Name, u.Phone, u.Role, &u.ID,
	).Err()
}

func (r *UserRepo) GetByEmail(email string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id,email,password_hash,name,phone,role,created_at FROM users WHERE email=:1`,
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Phone, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetByID(id int64) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRow(
		`SELECT id,email,name,phone,role,created_at FROM users WHERE id=:1`, id,
	).Scan(&u.ID, &u.Email, &u.Name, &u.Phone, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) Update(u *model.User) error {
	_, err := r.db.Exec(
		`UPDATE users SET name=:1,phone=:2 WHERE id=:3`,
		u.Name, u.Phone, u.ID,
	)
	return err
}

func (r *UserRepo) UpdatePassword(id int64, hash string) error {
	_, err := r.db.Exec(`UPDATE users SET password_hash=:1 WHERE id=:2`, hash, id)
	return err
}

// Address methods
func (r *UserRepo) CreateAddress(a *model.Address) error {
	return r.db.QueryRow(
		`INSERT INTO addresses(user_id,city,street,house,apartment,zip_code,is_default)
		 VALUES(:1,:2,:3,:4,:5,:6,:7) RETURNING id INTO :8`,
		a.UserID, a.City, a.Street, a.House, a.Apartment, a.ZipCode,
		boolToInt(a.IsDefault), &a.ID,
	).Err()
}

func (r *UserRepo) GetAddresses(userID int64) ([]model.Address, error) {
	rows, err := r.db.Query(
		`SELECT id,user_id,city,street,house,apartment,zip_code,is_default
		 FROM addresses WHERE user_id=:1 ORDER BY is_default DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	addrs := make([]model.Address, 0)
	for rows.Next() {
		var a model.Address
		var def int
		if err := rows.Scan(&a.ID, &a.UserID, &a.City, &a.Street, &a.House,
			&a.Apartment, &a.ZipCode, &def); err != nil {
			return nil, err
		}
		a.IsDefault = def == 1
		addrs = append(addrs, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return addrs, nil
}

func (r *UserRepo) UpdateAddress(a *model.Address) error {
	_, err := r.db.Exec(
		`UPDATE addresses SET city=:1,street=:2,house=:3,apartment=:4,zip_code=:5,is_default=:6
		 WHERE id=:7 AND user_id=:8`,
		a.City, a.Street, a.House, a.Apartment, a.ZipCode, boolToInt(a.IsDefault), a.ID, a.UserID,
	)
	return err
}

func (r *UserRepo) DeleteAddress(id, userID int64) error {
	_, err := r.db.Exec(`DELETE FROM addresses WHERE id=:1 AND user_id=:2`, id, userID)
	return err
}

func (r *UserRepo) SetDefaultAddress(id, userID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.Exec(`UPDATE addresses SET is_default=0 WHERE user_id=:1`, userID); err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE addresses SET is_default=1 WHERE id=:1 AND user_id=:2`, id, userID); err != nil {
		return err
	}
	return tx.Commit()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
