package repository

import (
	"context"
	"testing"
)

func TestGetByChatID_NotFound(t *testing.T) {
	db := testDB(t)
	repo := NewUserRepository(db)

	user, err := repo.GetByChatID(context.Background(), 999999999)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if user != nil {
		t.Fatalf("expected nil user, got: %+v", user)
	}
}

func TestUserRepository_CreateAndGet(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	repo := NewUserRepository(db)

	// Получаем ID группы "5-6" (заполнена миграцией 001)
	groupID, err := repo.GetGroupIDByName(ctx, "5-6")
	if err != nil {
		t.Fatalf("GetGroupIDByName: %v", err)
	}
	if groupID == 0 {
		t.Fatal("expected non-zero group ID for '5-6'")
	}

	// Создаём пользователя
	chatID := int64(777000001)
	u := &User{
		TgChatID:     chatID,
		TgUsername:    "test_user",
		FullName:     "Тест Тестов",
		Phone:        "+79991234567",
		ClassNumber:  5,
		ClassGroupID: groupID,
	}
	id, err := repo.Create(ctx, u)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero user ID")
	}

	// Читаем обратно
	found, err := repo.GetByChatID(ctx, chatID)
	if err != nil {
		t.Fatalf("GetByChatID: %v", err)
	}
	if found == nil {
		t.Fatal("expected user, got nil")
	}
	if found.FullName != "Тест Тестов" {
		t.Fatalf("expected FullName 'Тест Тестов', got '%s'", found.FullName)
	}
	if found.ClassGroupID != groupID {
		t.Fatalf("expected ClassGroupID %d, got %d", groupID, found.ClassGroupID)
	}

	// Cleanup
	db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
}

func TestGetGroupIDByName_AllGroups(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	repo := NewUserRepository(db)

	for _, name := range []string{"5-6", "7-9", "10-11"} {
		id, err := repo.GetGroupIDByName(ctx, name)
		if err != nil {
			t.Fatalf("GetGroupIDByName(%q): %v", name, err)
		}
		if id == 0 {
			t.Fatalf("expected non-zero ID for group %q", name)
		}
	}
}
