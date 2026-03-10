package repository

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type AdminUser struct {
	ID           int       `db:"id"`
	Username     string    `db:"username"`
	PasswordHash string    `db:"password_hash"`
	CreatedAt    time.Time `db:"created_at"`
}

type AdminUserRepository struct {
	db *sqlx.DB
}

func NewAdminUserRepository(db *sqlx.DB) *AdminUserRepository {
	return &AdminUserRepository{db: db}
}

func (r *AdminUserRepository) GetByUsername(ctx context.Context, username string) (*AdminUser, error) {
	var u AdminUser
	err := r.db.GetContext(ctx, &u, `SELECT * FROM admin_users WHERE username = $1`, username)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AdminUserRepository) Create(ctx context.Context, username, password string) (*AdminUser, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var u AdminUser
	err = r.db.QueryRowxContext(ctx,
		`INSERT INTO admin_users (username, password_hash) VALUES ($1, $2) RETURNING *`,
		username, string(hash),
	).StructScan(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AdminUserRepository) CheckPassword(user *AdminUser, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) == nil
}
