package repository

import (
	"context"
	"testing"
)

func TestTutorRepository_CRUD(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	repo := NewTutorRepository(db)

	// Create
	chatID := int64(111222333)
	tutor, err := repo.Create(ctx, "Иванов Иван", &chatID)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if tutor.ID == 0 {
		t.Fatal("expected non-zero tutor ID")
	}
	if tutor.FullName != "Иванов Иван" {
		t.Fatalf("expected 'Иванов Иван', got '%s'", tutor.FullName)
	}

	// GetByID
	found, err := repo.GetByID(ctx, tutor.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if found.FullName != "Иванов Иван" {
		t.Fatalf("expected 'Иванов Иван', got '%s'", found.FullName)
	}

	// GetAll
	all, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}
	foundInAll := false
	for _, tt := range all {
		if tt.ID == tutor.ID {
			foundInAll = true
			break
		}
	}
	if !foundInAll {
		t.Fatal("created tutor not found in GetAll")
	}

	// Update
	newChatID := int64(444555666)
	updated, err := repo.Update(ctx, tutor.ID, "Петров Пётр", &newChatID)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.FullName != "Петров Пётр" {
		t.Fatalf("expected 'Петров Пётр', got '%s'", updated.FullName)
	}

	// Create без chat ID (nil)
	tutorNoChatID, err := repo.Create(ctx, "Без Телеграма", nil)
	if err != nil {
		t.Fatalf("Create without chat ID: %v", err)
	}
	if tutorNoChatID.TgChatID != nil {
		t.Fatalf("expected nil chat ID, got %v", *tutorNoChatID.TgChatID)
	}

	// Delete
	if err := repo.Delete(ctx, tutor.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := repo.Delete(ctx, tutorNoChatID.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Проверяем, что удалился
	_, err = repo.GetByID(ctx, tutor.ID)
	if err == nil {
		t.Fatal("expected error after delete, got nil")
	}
}

func TestTutorRepository_SetSubjectsAndGroups(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	repo := NewTutorRepository(db)

	// Создаём репетитора
	tutor, err := repo.Create(ctx, "Тест Предметов", nil)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer repo.Delete(ctx, tutor.ID)

	// Получаем ID существующих предметов (из миграции 002)
	var subjectIDs []int
	err = db.SelectContext(ctx, &subjectIDs, `SELECT id FROM subjects ORDER BY id LIMIT 2`)
	if err != nil || len(subjectIDs) < 2 {
		t.Fatalf("need at least 2 subjects, got %d: %v", len(subjectIDs), err)
	}

	// SetSubjects
	if err := repo.SetSubjects(ctx, tutor.ID, subjectIDs); err != nil {
		t.Fatalf("SetSubjects: %v", err)
	}
	gotSubjects, err := repo.GetSubjectIDs(ctx, tutor.ID)
	if err != nil {
		t.Fatalf("GetSubjectIDs: %v", err)
	}
	if len(gotSubjects) != len(subjectIDs) {
		t.Fatalf("expected %d subjects, got %d", len(subjectIDs), len(gotSubjects))
	}

	// Получаем ID существующих групп (из миграции 001)
	var groupIDs []int
	err = db.SelectContext(ctx, &groupIDs, `SELECT id FROM class_groups ORDER BY id LIMIT 2`)
	if err != nil || len(groupIDs) < 2 {
		t.Fatalf("need at least 2 groups, got %d: %v", len(groupIDs), err)
	}

	// SetGroups
	if err := repo.SetGroups(ctx, tutor.ID, groupIDs); err != nil {
		t.Fatalf("SetGroups: %v", err)
	}
	gotGroups, err := repo.GetGroupIDs(ctx, tutor.ID)
	if err != nil {
		t.Fatalf("GetGroupIDs: %v", err)
	}
	if len(gotGroups) != len(groupIDs) {
		t.Fatalf("expected %d groups, got %d", len(groupIDs), len(gotGroups))
	}

	// Перезаписываем на один предмет
	if err := repo.SetSubjects(ctx, tutor.ID, subjectIDs[:1]); err != nil {
		t.Fatalf("SetSubjects (overwrite): %v", err)
	}
	gotSubjects, _ = repo.GetSubjectIDs(ctx, tutor.ID)
	if len(gotSubjects) != 1 {
		t.Fatalf("expected 1 subject after overwrite, got %d", len(gotSubjects))
	}
}
