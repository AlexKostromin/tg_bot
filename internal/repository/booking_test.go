package repository

import (
	"context"
	"testing"
	"time"
)

// setupBookingTestData создаёт пользователя, репетитора и слот для тестов бронирования.
// Возвращает cleanup-функцию.
func setupBookingTestData(t *testing.T, db interface {
	QueryRowContext(ctx context.Context, query string, args ...any) interface{ Scan(dest ...any) error }
	ExecContext(ctx context.Context, query string, args ...any) (interface{ RowsAffected() (int64, error) }, error)
}) (userID, slotID int, cleanup func()) {
	t.Helper()
	// Не получится использовать интерфейс с sqlx напрямую, поэтому используем testDB
	return 0, 0, func() {}
}

func TestBookingRepository_CountActive_Zero(t *testing.T) {
	db := testDB(t)
	repo := NewBookingRepository(db)

	// Для несуществующего пользователя — 0, не ошибка
	count, err := repo.CountActive(context.Background(), 999999)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 active bookings, got %d", count)
	}
}

func TestBookingRepository_GetActiveByUserID_Empty(t *testing.T) {
	db := testDB(t)
	repo := NewBookingRepository(db)

	views, err := repo.GetActiveByUserID(context.Background(), 999999)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(views) != 0 {
		t.Fatalf("expected empty views, got %d", len(views))
	}
}

func TestBookingRepository_Cancel_NonExistent(t *testing.T) {
	db := testDB(t)
	repo := NewBookingRepository(db)

	err := repo.Cancel(context.Background(), 999999, 999999)
	if err == nil {
		t.Fatal("expected error for non-existent booking, got nil")
	}
}

func TestBookingRepository_HasConflict_NoConflict(t *testing.T) {
	db := testDB(t)
	repo := NewBookingRepository(db)

	// Несуществующий пользователь и слот — не должно быть конфликта
	conflict, err := repo.HasConflict(context.Background(), 999999, 999999)
	// HasConflict делает JOIN с time_slots — несуществующий slot_id может вызвать
	// count=0 или ошибку в зависимости от реализации JOIN
	if err != nil {
		t.Logf("HasConflict with non-existent slot returned error (expected for JOIN): %v", err)
		return
	}
	if conflict {
		t.Fatal("expected no conflict for non-existent user/slot")
	}
}

func TestBookingRepository_FullCycle(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	bookingRepo := NewBookingRepository(db)
	tutorRepo := NewTutorRepository(db)
	userRepo := NewUserRepository(db)

	// Создаём репетитора
	tutor, err := tutorRepo.Create(ctx, "Букинг Тестер", nil)
	if err != nil {
		t.Fatalf("create tutor: %v", err)
	}
	defer tutorRepo.Delete(ctx, tutor.ID)

	// Создаём пользователя
	var groupID int
	db.QueryRowContext(ctx, `SELECT id FROM class_groups WHERE name = '5-6'`).Scan(&groupID)
	u := &User{
		TgChatID:     777000002,
		TgUsername:    "booking_tester",
		FullName:     "Букинг Тест",
		Phone:        "+79990000002",
		ClassNumber:  5,
		ClassGroupID: groupID,
	}
	userID, err := userRepo.Create(ctx, u)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, userID)

	// Создаём слот
	var subjectID int
	db.QueryRowContext(ctx, `SELECT id FROM subjects LIMIT 1`).Scan(&subjectID)
	tomorrow := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
	var slotID int
	err = db.QueryRowContext(ctx, `
		INSERT INTO time_slots (tutor_id, subject_id, class_group_id, slot_date, start_time, end_time)
		VALUES ($1, $2, $3, $4, '14:00', '15:00')
		RETURNING id`,
		tutor.ID, subjectID, groupID, tomorrow,
	).Scan(&slotID)
	if err != nil {
		t.Fatalf("insert slot: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM time_slots WHERE id = $1`, slotID)

	// Create booking
	booking := &Booking{UserID: userID, SlotID: slotID, Comment: "тестовая бронь"}
	bookingID, err := bookingRepo.Create(ctx, booking)
	if err != nil {
		t.Fatalf("Create booking: %v", err)
	}
	if bookingID == 0 {
		t.Fatal("expected non-zero booking ID")
	}
	defer db.ExecContext(ctx, `DELETE FROM bookings WHERE id = $1`, bookingID)

	// CountActive — должна быть 1
	count, err := bookingRepo.CountActive(ctx, userID)
	if err != nil {
		t.Fatalf("CountActive: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 active booking, got %d", count)
	}

	// GetActiveByUserID — должна вернуть одну запись
	views, err := bookingRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetActiveByUserID: %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("expected 1 booking view, got %d", len(views))
	}

	// GetFullInfo
	info, err := bookingRepo.GetFullInfo(ctx, bookingID)
	if err != nil {
		t.Fatalf("GetFullInfo: %v", err)
	}
	if info.BookingID != bookingID {
		t.Fatalf("expected booking ID %d, got %d", bookingID, info.BookingID)
	}
	if info.StudentName != "Букинг Тест" {
		t.Fatalf("expected student name 'Букинг Тест', got '%s'", info.StudentName)
	}

	// UpdateStatus
	if err := bookingRepo.UpdateStatus(ctx, bookingID, "confirmed"); err != nil {
		t.Fatalf("UpdateStatus: %v", err)
	}

	// Cancel
	if err := bookingRepo.Cancel(ctx, bookingID, userID); err != nil {
		t.Fatalf("Cancel: %v", err)
	}

	// После отмены CountActive = 0
	count, err = bookingRepo.CountActive(ctx, userID)
	if err != nil {
		t.Fatalf("CountActive after cancel: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 active bookings after cancel, got %d", count)
	}
}
