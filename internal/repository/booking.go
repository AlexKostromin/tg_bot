package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// internal/repository/booking.go

type Booking struct {
	ID      int    `db:"id"`
	UserID  int    `db:"user_id"`
	SlotID  int    `db:"slot_id"`
	Status  string `db:"status"`
	Comment string `db:"comment"`
}

// FullInfo — полная информация о брони для уведомлений.
// Один JOIN-запрос вместо нескольких отдельных вызовов.
type FullInfo struct {
	BookingID     int       `db:"booking_id"`
	UserID        int       `db:"user_id"`
	Status        string    `db:"status"`
	StudentName   string    `db:"student_name"`
	StudentChatID int64     `db:"student_chat_id"`
	Phone         string    `db:"phone"`
	ClassNumber   int       `db:"class_number"`
	SubjectID     int       `db:"subject_id"`
	SubjectName   string    `db:"subject_name"`
	SlotDate      time.Time `db:"slot_date"`
	StartTime     string    `db:"start_time"`
	EndTime       string    `db:"end_time"`
	TutorChatID   *int64    `db:"tutor_chat_id"`
	BookedAt      time.Time `db:"booked_at"`
	SlotID        int       `db:"slot_id"`
}

type BookingRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// BookingView — расширенная структура для отображения записей пользователю.
// Объединяет данные из bookings, time_slots и subjects через JOIN.
type BookingView struct {
	ID          int       `db:"id"`
	Status      string    `db:"status"`
	SubjectName string    `db:"subject_name"`
	SlotDate    time.Time `db:"slot_date"`
	StartTime   string    `db:"start_time"`
	EndTime     string    `db:"end_time"`
	BookedAt    time.Time `db:"booked_at"`
}

// Create выполняется в транзакции: INSERT в bookings атомарен
// с UPDATE time_slots через триггер. При откате — оба изменения отменятся.
func (r *BookingRepository) Create(ctx context.Context, b *Booking) (int, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() // no-op если был Commit()

	var id int
	err = tx.QueryRowContext(ctx, `
        INSERT INTO bookings (user_id, slot_id, status, comment)
        VALUES ($1, $2, 'pending', $3)
        RETURNING id`,
		b.UserID, b.SlotID, b.Comment,
	).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit()
}

// GetActiveByUserID показывает предстоящие записи.
// Фильтр slot_date >= CURRENT_DATE исключает прошедшие занятия.
func (r *BookingRepository) GetActiveByUserID(ctx context.Context, userID int) ([]BookingView, error) {
	var views []BookingView
	err := r.db.SelectContext(ctx, &views, `
        SELECT b.id, b.status, b.booked_at,
               s.name AS subject_name,
               ts.slot_date, ts.start_time, ts.end_time
        FROM bookings b
        JOIN time_slots ts ON ts.id = b.slot_id
        JOIN subjects   s  ON s.id  = ts.subject_id
        WHERE b.user_id = $1
          AND b.status IN ('pending', 'confirmed')
          AND ts.slot_date >= CURRENT_DATE
        ORDER BY ts.slot_date, ts.start_time`,
		userID,
	)
	return views, err
}

// Cancel меняет статус. Триггер trg_booking_cancelled автоматически
// освободит слот. RowsAffected() == 0 — запись не найдена или уже отменена.
func (r *BookingRepository) Cancel(ctx context.Context, bookingID, userID int) error {
	result, err := r.db.ExecContext(ctx, `
        UPDATE bookings
        SET status = 'cancelled'
        WHERE id      = $1
          AND user_id = $2
          AND status IN ('pending', 'confirmed')`,
		bookingID, userID,
	)
	if err != nil {
		return err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return errors.New("запись не найдена или уже отменена")
	}
	return nil
}

// HasConflict проверяет пересечение по времени с существующей записью.
func (r *BookingRepository) HasConflict(ctx context.Context, userID, slotID int) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM bookings b
        JOIN time_slots new_slot ON new_slot.id = $2
        JOIN time_slots old_slot ON old_slot.id = b.slot_id
        WHERE b.user_id = $1
          AND b.status IN ('pending', 'confirmed')
          AND old_slot.slot_date = new_slot.slot_date
          AND old_slot.start_time < new_slot.end_time
          AND old_slot.end_time   > new_slot.start_time`,
		userID, slotID,
	).Scan(&count)
	return count > 0, err
}

// GetFullInfo — полная информация о брони для уведомлений.
// Один JOIN-запрос вместо нескольких отдельных вызовов.
func (r *BookingRepository) GetFullInfo(ctx context.Context, bookingID int) (*FullInfo, error) {
	var info FullInfo
	err := r.db.GetContext(ctx, &info, `
        SELECT b.id AS booking_id,
               b.user_id,
               b.slot_id,
               b.status,
               b.booked_at,
               u.full_name AS student_name, u.tg_chat_id AS student_chat_id,
               u.phone, u.class_number,
               s.id AS subject_id, s.name AS subject_name,
               ts.slot_date, ts.start_time, ts.end_time,
               t.tg_chat_id AS tutor_chat_id
        FROM bookings b
        JOIN users      u  ON u.id  = b.user_id
        JOIN time_slots ts ON ts.id = b.slot_id
        JOIN subjects   s  ON s.id  = ts.subject_id
        JOIN tutors     t  ON t.id  = ts.tutor_id
        WHERE b.id = $1`,
		bookingID,
	)
	return &info, err
}

// CountActive — количество активных записей пользователя.
// Используется для проверки лимита MaxActiveBookings.
func (r *BookingRepository) CountActive(ctx context.Context, userID int) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM bookings
        WHERE user_id = $1
          AND status IN ('pending', 'confirmed')`,
		userID,
	).Scan(&count)
	return count, err
}

// UpdateStatus меняет статус брони. Используется админом для подтверждения/отклонения.
func (r *BookingRepository) UpdateStatus(ctx context.Context, bookingID int, status string) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE bookings SET status = $1 WHERE id = $2`,
		status, bookingID,
	)
	return err
}

// --- Admin API types and methods ---

type BookingFilters struct {
	Status string
	UserID int
	Date   string
	Offset int
	Limit  int
}

// ListWithFilters возвращает брони с фильтрами для админки.
func (r *BookingRepository) ListWithFilters(ctx context.Context, f BookingFilters) ([]FullInfo, int, error) {
	query := `
		SELECT b.id AS booking_id,
		       b.user_id,
		       b.slot_id,
		       b.status,
		       b.booked_at,
		       u.full_name AS student_name, u.tg_chat_id AS student_chat_id,
		       u.phone, u.class_number,
		       s.id AS subject_id, s.name AS subject_name,
		       ts.slot_date, ts.start_time, ts.end_time,
		       t.tg_chat_id AS tutor_chat_id
		FROM bookings b
		JOIN users      u  ON u.id  = b.user_id
		JOIN time_slots ts ON ts.id = b.slot_id
		JOIN subjects   s  ON s.id  = ts.subject_id
		JOIN tutors     t  ON t.id  = ts.tutor_id
		WHERE 1=1`
	countQuery := `SELECT COUNT(*) FROM bookings b JOIN time_slots ts ON ts.id = b.slot_id WHERE 1=1`
	var args []interface{}
	idx := 1

	if f.Status != "" {
		query += fmt.Sprintf(` AND b.status = $%d`, idx)
		countQuery += fmt.Sprintf(` AND b.status = $%d`, idx)
		args = append(args, f.Status)
		idx++
	}
	if f.UserID > 0 {
		query += fmt.Sprintf(` AND b.user_id = $%d`, idx)
		countQuery += fmt.Sprintf(` AND b.user_id = $%d`, idx)
		args = append(args, f.UserID)
		idx++
	}
	if f.Date != "" {
		query += fmt.Sprintf(` AND ts.slot_date = $%d`, idx)
		countQuery += fmt.Sprintf(` AND ts.slot_date = $%d`, idx)
		args = append(args, f.Date)
		idx++
	}

	var total int
	r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)

	query += fmt.Sprintf(` ORDER BY ts.slot_date DESC, ts.start_time LIMIT $%d OFFSET $%d`, idx, idx+1)
	args = append(args, f.Limit, f.Offset)

	var items []FullInfo
	err := r.db.SelectContext(ctx, &items, query, args...)
	return items, total, err
}
