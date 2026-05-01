package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const selectCols = `id, email, full_name, username, avatar_url, provider, password_hash, created_at, updated_at`

func (r *Repository) scan(row pgx.Row) (*User, error) {
	u := &User{}
	err := row.Scan(&u.ID, &u.Email, &u.FullName, &u.Username, &u.AvatarURL, &u.Provider, &u.PasswordHash, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return u, err
}

func (r *Repository) GetByID(ctx context.Context, id string) (*User, error) {
	return r.scan(r.db.QueryRow(ctx,
		`SELECT `+selectCols+` FROM users WHERE id = $1`, id,
	))
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	return r.scan(r.db.QueryRow(ctx,
		`SELECT `+selectCols+` FROM users WHERE email = $1`, email,
	))
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (*User, error) {
	return r.scan(r.db.QueryRow(ctx,
		`SELECT `+selectCols+` FROM users WHERE username = $1`, username,
	))
}

func (r *Repository) Create(ctx context.Context, u *User) (*User, error) {
	return r.scan(r.db.QueryRow(ctx,
		`INSERT INTO users (email, full_name, username, avatar_url, provider, password_hash)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING `+selectCols,
		u.Email, u.FullName, u.Username, u.AvatarURL, u.Provider, u.PasswordHash,
	))
}

// UpsertByEmail upserts a social-login user: on conflict updates name and avatar.
func (r *Repository) UpsertByEmail(ctx context.Context, u *User) (*User, error) {
	return r.scan(r.db.QueryRow(ctx,
		`INSERT INTO users (email, full_name, username, avatar_url, provider)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (email) DO UPDATE
		   SET full_name  = COALESCE(EXCLUDED.full_name, users.full_name),
		       avatar_url = COALESCE(EXCLUDED.avatar_url, users.avatar_url),
		       updated_at = NOW()
		 RETURNING `+selectCols,
		u.Email, u.FullName, u.Username, u.AvatarURL, u.Provider,
	))
}
