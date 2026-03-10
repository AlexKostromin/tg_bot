package repository

import (
	"context"
	"testing"
)

func TestSubjectRepository_GetByGroupID(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	repo := NewSubjectRepository(db)

	// Группа "5-6" существует (миграция 001), предметы привязаны (миграция 002)
	var groupID int
	err := db.QueryRowContext(ctx, `SELECT id FROM class_groups WHERE name = '5-6'`).Scan(&groupID)
	if err != nil {
		t.Fatalf("get group ID: %v", err)
	}

	subjects, err := repo.GetByGroupID(ctx, groupID)
	if err != nil {
		t.Fatalf("GetByGroupID: %v", err)
	}
	if len(subjects) == 0 {
		t.Fatal("expected at least one subject for group '5-6'")
	}

	// Проверяем, что у каждого предмета заполнены поля
	for _, s := range subjects {
		if s.ID == 0 || s.Name == "" {
			t.Fatalf("invalid subject: %+v", s)
		}
	}
	t.Logf("OK: found %d subjects for group '5-6'", len(subjects))
}

func TestSubjectRepository_GetByGroupID_NonExistent(t *testing.T) {
	db := testDB(t)
	repo := NewSubjectRepository(db)

	subjects, err := repo.GetByGroupID(context.Background(), 999999)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if len(subjects) != 0 {
		t.Fatalf("expected empty slice, got %d subjects", len(subjects))
	}
}
