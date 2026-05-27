package auth

import (
	"context"
	"database/sql"
	"errors"
)

// Repository handles users-table reads/writes required by auth (and exposes
// helpers used by other modules to fetch a user by id).
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

// FindByPhone returns the user matching the phone number or sql.ErrNoRows.
func (r *Repository) FindByPhone(ctx context.Context, phone string) (*User, error) {
	const q = `
		SELECT id, name, phone_number, email, role, created_at, updated_at
		  FROM users
		 WHERE phone_number = $1`
	u := &User{}
	err := r.db.QueryRowContext(ctx, q, phone).Scan(
		&u.ID, &u.Name, &u.PhoneNumber, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// FindByID returns a user by id.
func (r *Repository) FindByID(ctx context.Context, id string) (*User, error) {
	const q = `
		SELECT id, name, phone_number, email, role, created_at, updated_at
		  FROM users
		 WHERE id = $1`
	u := &User{}
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&u.ID, &u.Name, &u.PhoneNumber, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// CreateCustomer inserts a customer user. Used on first OTP verification.
func (r *Repository) CreateCustomer(ctx context.Context, phone string) (*User, error) {
	const q = `
		INSERT INTO users (name, phone_number, role)
		VALUES ('', $1, 'customer')
		RETURNING id, name, phone_number, email, role, created_at, updated_at`
	u := &User{}
	err := r.db.QueryRowContext(ctx, q, phone).Scan(
		&u.ID, &u.Name, &u.PhoneNumber, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	return u, err
}

// EnsureWallet creates a wallet row for the user if one does not exist.
// Called on first login so debits/credits never have to bootstrap the row.
func (r *Repository) EnsureWallet(ctx context.Context, userID string) error {
	const q = `INSERT INTO wallets (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING`
	_, err := r.db.ExecContext(ctx, q, userID)
	return err
}

// UpdateProfile updates name and/or email.
func (r *Repository) UpdateProfile(ctx context.Context, userID, name string, email *string) (*User, error) {
	const q = `
		UPDATE users
		   SET name       = COALESCE(NULLIF($2, ''), name),
		       email      = COALESCE($3, email),
		       updated_at = now()
		 WHERE id = $1
		RETURNING id, name, phone_number, email, role, created_at, updated_at`
	u := &User{}
	err := r.db.QueryRowContext(ctx, q, userID, name, email).Scan(
		&u.ID, &u.Name, &u.PhoneNumber, &u.Email, &u.Role, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	return u, err
}
