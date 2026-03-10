package fsm

import (
	"context"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
)

func testRedis(t *testing.T) *redis.Client {
	t.Helper()
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASSWORD"),
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis not available: %v", err)
	}
	t.Cleanup(func() { rdb.Close() })
	return rdb
}

func TestStorage_SetAndGet(t *testing.T) {
	rdb := testRedis(t)
	s := NewStorage(rdb)
	ctx := context.Background()
	chatID := int64(100500)
	defer s.Clear(ctx, chatID)

	// Set
	us := &UserState{
		State: StateRegAwaitName,
		Data:  map[string]string{"name": "Тест", "phone": "+7999"},
	}
	if err := s.Set(ctx, chatID, us); err != nil {
		t.Fatalf("Set: %v", err)
	}

	// Get — проверяем, что State и Data совпадают
	got, err := s.Get(ctx, chatID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.State != StateRegAwaitName {
		t.Fatalf("expected state %q, got %q", StateRegAwaitName, got.State)
	}
	if got.Data["name"] != "Тест" {
		t.Fatalf("expected Data[name]='Тест', got %q", got.Data["name"])
	}
	if got.Data["phone"] != "+7999" {
		t.Fatalf("expected Data[phone]='+7999', got %q", got.Data["phone"])
	}
}

func TestStorage_Get_NoKey(t *testing.T) {
	rdb := testRedis(t)
	s := NewStorage(rdb)

	// Несуществующий ключ — StateNone, пустая Data, без ошибки
	got, err := s.Get(context.Background(), 999888777)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if got.State != StateNone {
		t.Fatalf("expected StateNone, got %q", got.State)
	}
	if len(got.Data) != 0 {
		t.Fatalf("expected empty Data, got %v", got.Data)
	}
}

func TestStorage_Transition(t *testing.T) {
	rdb := testRedis(t)
	s := NewStorage(rdb)
	ctx := context.Background()
	chatID := int64(100501)
	defer s.Clear(ctx, chatID)

	// Начальное состояние с данными
	s.Set(ctx, chatID, &UserState{
		State: StateRegAwaitName,
		Data:  map[string]string{"name": "Алекс"},
	})

	// Transition меняет State, но сохраняет Data
	if err := s.Transition(ctx, chatID, StateRegAwaitPhone); err != nil {
		t.Fatalf("Transition: %v", err)
	}
	got, _ := s.Get(ctx, chatID)
	if got.State != StateRegAwaitPhone {
		t.Fatalf("expected %q, got %q", StateRegAwaitPhone, got.State)
	}
	if got.Data["name"] != "Алекс" {
		t.Fatalf("Data lost after Transition: %v", got.Data)
	}
}

func TestStorage_TransitionWithData(t *testing.T) {
	rdb := testRedis(t)
	s := NewStorage(rdb)
	ctx := context.Background()
	chatID := int64(100502)
	defer s.Clear(ctx, chatID)

	s.Set(ctx, chatID, &UserState{
		State: StateBookAwaitSubject,
		Data:  map[string]string{"old_key": "old_val"},
	})

	// TransitionWithData меняет State и добавляет поле
	if err := s.TransitionWithData(ctx, chatID, StateBookAwaitDate, "subject_id", "5"); err != nil {
		t.Fatalf("TransitionWithData: %v", err)
	}
	got, _ := s.Get(ctx, chatID)
	if got.State != StateBookAwaitDate {
		t.Fatalf("expected %q, got %q", StateBookAwaitDate, got.State)
	}
	if got.Data["subject_id"] != "5" {
		t.Fatalf("expected subject_id='5', got %q", got.Data["subject_id"])
	}
	if got.Data["old_key"] != "old_val" {
		t.Fatalf("old data lost: %v", got.Data)
	}
}

func TestStorage_SetField(t *testing.T) {
	rdb := testRedis(t)
	s := NewStorage(rdb)
	ctx := context.Background()
	chatID := int64(100503)
	defer s.Clear(ctx, chatID)

	s.Set(ctx, chatID, &UserState{State: StateRegAwaitClass, Data: map[string]string{}})

	if err := s.SetField(ctx, chatID, "class", "7"); err != nil {
		t.Fatalf("SetField: %v", err)
	}
	got, _ := s.Get(ctx, chatID)
	if got.Data["class"] != "7" {
		t.Fatalf("expected class='7', got %q", got.Data["class"])
	}
	// State не изменился
	if got.State != StateRegAwaitClass {
		t.Fatalf("State changed unexpectedly: %q", got.State)
	}
}

func TestStorage_Clear(t *testing.T) {
	rdb := testRedis(t)
	s := NewStorage(rdb)
	ctx := context.Background()
	chatID := int64(100504)

	s.Set(ctx, chatID, &UserState{State: StateBookConfirm, Data: map[string]string{"x": "y"}})

	if err := s.Clear(ctx, chatID); err != nil {
		t.Fatalf("Clear: %v", err)
	}

	// После Clear — StateNone
	got, _ := s.Get(ctx, chatID)
	if got.State != StateNone {
		t.Fatalf("expected StateNone after Clear, got %q", got.State)
	}
}
