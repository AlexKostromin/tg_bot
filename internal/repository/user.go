package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

type User struct {
	ID           int       `db:"id"`
	TgChatID     int64     `db:"tg_chat_id"`
	TgUsername   string    `db:"tg_username"`
	FullName     string    `db:"full_name"`
	Phone        string    `db:"phone"`
	ClassNumber  int       `db:"class_number"`
	ClassGroupID int       `db:"class_group_id"`
	IsActive     bool      `db:"is_active"`
	RegisteredAt time.Time `db:"registered_at"`
}

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByChatID ищет пользователя по Telegram chat ID.
// Вызывается при каждом входящем сообщении, чтобы понять —
// зарегистрирован пользователь или нет.
// Возвращает (nil, nil), если не найден — это не ошибка.
func (r *UserRepository) GetByChatID(ctx context.Context, chatID int64) (*User, error) {
	var u User
	err := r.db.GetContext(ctx, &u, "SELECT * FROM users WHERE tg_chat_id = $1", chatID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return &u, err
}

// Create вставляет нового пользователя в базу после прохождения регистрации.
// Возвращает сгенерированный ID через RETURNING.
func (r *UserRepository) Create(ctx context.Context, u *User) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `INSERT INTO users (tg_chat_id, tg_username, full_name, phone, class_number, class_group_id)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id`, u.TgChatID, u.TgUsername, u.FullName, u.Phone, u.ClassNumber, u.ClassGroupID).Scan(&id)
	return id, err
}
// GetGroupIDByName возвращает ID группы классов по её имени ("5-6", "7-9", "10-11").
// Используется при регистрации: сервис определяет имя группы по номеру класса,
// а этот метод переводит имя в ID для записи в users.class_group_id.
func (r *UserRepository) GetGroupIDByName(ctx context.Context, name string) (int, error) {
	var id int
	err := r.db.QueryRowContext(ctx, `SELECT id from class_groups WHERE name = $1`, name).Scan(&id)
	return id, err
}

// GetByID возвращает пользователя по внутреннему ID.
func (r *UserRepository) GetByID(ctx context.Context, id int) (*User, error) {
	var u User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// ListWithSearch возвращает пользователей с поиском по имени/телефону и пагинацией.
func (r *UserRepository) ListWithSearch(ctx context.Context, search string, offset, limit int) ([]User, int, error) {
	var total int
	var users []User

	if search != "" {
		pattern := "%" + search + "%"
		r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM users WHERE full_name ILIKE $1 OR phone ILIKE $1`, pattern).Scan(&total)
		err := r.db.SelectContext(ctx, &users,
			`SELECT * FROM users WHERE full_name ILIKE $1 OR phone ILIKE $1
			 ORDER BY registered_at DESC LIMIT $2 OFFSET $3`, pattern, limit, offset)
		return users, total, err
	}

	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&total)
	err := r.db.SelectContext(ctx, &users,
		`SELECT * FROM users ORDER BY registered_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	return users, total, err
}

// SetActive включает/выключает пользователя (блокировка из админки).
func (r *UserRepository) SetActive(ctx context.Context, id int, active bool) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET is_active = $1 WHERE id = $2`, active, id)
	return err
}
