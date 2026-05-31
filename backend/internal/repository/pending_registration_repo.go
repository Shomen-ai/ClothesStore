package repository

import (
	"database/sql"
	"errors"
	"time"
)

type PendingRegistration struct {
	Email        string
	Code         string
	PasswordHash string
	Name         string
	Phone        string
	Attempts     int
	ExpiresAt    time.Time
	ResendAfter  time.Time
}

type PendingRegistrationRepo struct{ db *sql.DB }

func NewPendingRegistrationRepo(db *sql.DB) *PendingRegistrationRepo {
	return &PendingRegistrationRepo{db: db}
}

// Upsert writes a fresh pending row, resetting attempts and timers — used both
// for the initial register/start and for register/resend.
func (r *PendingRegistrationRepo) Upsert(p *PendingRegistration) error {
	_, err := r.db.Exec(`
		INSERT INTO pending_registrations
		  (email, code, password_hash, name, phone, attempts, expires_at, resend_after)
		VALUES ($1, $2, $3, $4, $5, 0, $6, $7)
		ON CONFLICT (email) DO UPDATE SET
		  code          = EXCLUDED.code,
		  password_hash = EXCLUDED.password_hash,
		  name          = EXCLUDED.name,
		  phone         = EXCLUDED.phone,
		  attempts      = 0,
		  expires_at    = EXCLUDED.expires_at,
		  resend_after  = EXCLUDED.resend_after`,
		p.Email, p.Code, p.PasswordHash, p.Name, p.Phone, p.ExpiresAt, p.ResendAfter)
	return err
}

func (r *PendingRegistrationRepo) Get(email string) (*PendingRegistration, error) {
	var p PendingRegistration
	err := r.db.QueryRow(`
		SELECT email, code, password_hash, name, phone, attempts, expires_at, resend_after
		FROM pending_registrations WHERE email = $1`, email).
		Scan(&p.Email, &p.Code, &p.PasswordHash, &p.Name, &p.Phone, &p.Attempts, &p.ExpiresAt, &p.ResendAfter)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PendingRegistrationRepo) BumpAttempts(email string) error {
	_, err := r.db.Exec(`UPDATE pending_registrations SET attempts = attempts + 1 WHERE email = $1`, email)
	return err
}

func (r *PendingRegistrationRepo) Delete(email string) error {
	_, err := r.db.Exec(`DELETE FROM pending_registrations WHERE email = $1`, email)
	return err
}

// CleanExpired removes rows whose expiry passed more than `keepAfterExpire`
// ago, so verify can still respond "expired" for a recently-expired row.
func (r *PendingRegistrationRepo) CleanExpired(keepAfterExpire time.Duration) (int64, error) {
	res, err := r.db.Exec(
		`DELETE FROM pending_registrations WHERE expires_at < NOW() - ($1 || ' seconds')::interval`,
		int64(keepAfterExpire.Seconds()),
	)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}
