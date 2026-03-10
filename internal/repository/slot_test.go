package repository

import (
	"context"
	"testing"
	"time"
)

func TestSlotRepository_GetAvailableDates_Empty(t *testing.T) {
	db := testDB(t)
	repo := NewSlotRepository(db)

	// Для несуществующей группы/предмета — пустой слайс, не ошибка
	dates, err := repo.GetAvailableDates(context.Background(), 999999, 999999)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(dates) != 0 {
		t.Fatalf("expected empty dates, got %d", len(dates))
	}
}

func TestSlotRepository_GetAvailableSlots_Empty(t *testing.T) {
	db := testDB(t)
	repo := NewSlotRepository(db)

	slots, err := repo.GetAvailableSlots(context.Background(), 999999, 999999, time.Now())
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(slots) != 0 {
		t.Fatalf("expected empty slots, got %d", len(slots))
	}
}

func TestSlotRepository_GetByID_NotFound(t *testing.T) {
	db := testDB(t)
	repo := NewSlotRepository(db)

	_, err := repo.GetByID(context.Background(), 999999)
	if err == nil {
		t.Fatal("expected error for non-existent slot, got nil")
	}
}

func TestSlotRepository_CreateAndGet(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	slotRepo := NewSlotRepository(db)
	tutorRepo := NewTutorRepository(db)

	// Создаём репетитора для FK
	tutor, err := tutorRepo.Create(ctx, "Слот Тестер", nil)
	if err != nil {
		t.Fatalf("create tutor: %v", err)
	}
	defer tutorRepo.Delete(ctx, tutor.ID)

	// Берём ID предмета и группы из миграций
	var subjectID, groupID int
	db.QueryRowContext(ctx, `SELECT id FROM subjects LIMIT 1`).Scan(&subjectID)
	db.QueryRowContext(ctx, `SELECT id FROM class_groups LIMIT 1`).Scan(&groupID)
	if subjectID == 0 || groupID == 0 {
		t.Fatal("need seed data: subjects and class_groups")
	}

	// Вставляем слот напрямую (у SlotRepo пока нет Create)
	tomorrow := time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour)
	var slotID int
	err = db.QueryRowContext(ctx, `
		INSERT INTO time_slots (tutor_id, subject_id, class_group_id, slot_date, start_time, end_time)
		VALUES ($1, $2, $3, $4, '10:00', '11:00')
		RETURNING id`,
		tutor.ID, subjectID, groupID, tomorrow,
	).Scan(&slotID)
	if err != nil {
		t.Fatalf("insert slot: %v", err)
	}
	defer db.ExecContext(ctx, `DELETE FROM time_slots WHERE id = $1`, slotID)

	// GetByID
	slot, err := slotRepo.GetByID(ctx, slotID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if slot.ID != slotID {
		t.Fatalf("expected slot ID %d, got %d", slotID, slot.ID)
	}

	// GetAvailableDates — должна содержать tomorrow
	dates, err := slotRepo.GetAvailableDates(ctx, groupID, subjectID)
	if err != nil {
		t.Fatalf("GetAvailableDates: %v", err)
	}
	found := false
	for _, d := range dates {
		if d.Truncate(24*time.Hour).Equal(tomorrow) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected tomorrow %s in available dates, got %v", tomorrow.Format("2006-01-02"), dates)
	}

	// GetAvailableSlots — должен содержать наш слот
	slots, err := slotRepo.GetAvailableSlots(ctx, groupID, subjectID, tomorrow)
	if err != nil {
		t.Fatalf("GetAvailableSlots: %v", err)
	}
	if len(slots) == 0 {
		t.Fatal("expected at least one slot")
	}
}
