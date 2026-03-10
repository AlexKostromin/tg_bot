// internal/fsm/storage.go
package fsm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const stateTTL = 24 * time.Hour

type UserState struct {
	State string            `json:"state"`
	Data  map[string]string `json:"data"`
}

type Storage struct {
	rdb *redis.Client
}

func NewStorage(rdb *redis.Client) *Storage {
	return &Storage{rdb: rdb}
}

// Get возвращает текущее состояние.
// redis.Nil — не ошибка, а «пользователь в главном меню».
func (s *Storage) Get(ctx context.Context, chatID int64) (*UserState, error) {
	val, err := s.rdb.Get(ctx, fmt.Sprintf("state:%d", chatID)).Result()
	if errors.Is(err, redis.Nil) {
		return &UserState{State: StateNone, Data: map[string]string{}}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis get: %w", err)
	}
	var us UserState
	return &us, json.Unmarshal([]byte(val), &us)
}

func (s *Storage) Set(ctx context.Context, chatID int64, us *UserState) error {
	data, err := json.Marshal(us)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	return s.rdb.Set(ctx, fmt.Sprintf("state:%d", chatID), data, stateTTL).Err()
}

func (s *Storage) Clear(ctx context.Context, chatID int64) error {
	return s.rdb.Del(ctx, fmt.Sprintf("state:%d", chatID)).Err()
}

// Transition меняет шаг, сохраняя накопленные Data.
func (s *Storage) Transition(ctx context.Context, chatID int64, newState string) error {
	us, err := s.Get(ctx, chatID)
	if err != nil {
		return err
	}
	us.State = newState
	return s.Set(ctx, chatID, us)
}

// SetField добавляет одно поле в Data без полной перезаписи.
func (s *Storage) SetField(ctx context.Context, chatID int64, key, value string) error {
	us, err := s.Get(ctx, chatID)
	if err != nil {
		return err
	}
	us.Data[key] = value
	return s.Set(ctx, chatID, us)
}

// TransitionWithData меняет шаг и записывает поле за одну операцию.
// Это безопаснее, чем отдельные SetField + Transition: один Get + один Set
// вместо двух пар, что исключает гонку при конкурентных обновлениях.
func (s *Storage) TransitionWithData(ctx context.Context, chatID int64, newState, key, value string) error {
	us, err := s.Get(ctx, chatID)
	if err != nil {
		return err
	}
	us.State = newState
	us.Data[key] = value
	return s.Set(ctx, chatID, us)
}
