package service

import (
	"context"
	"fmt"

	"github.com/AlexKostromin/tg_bot/internal/repository"
)

// ClassGroupByNumber определяет группу классов по номеру класса ученика.
// Бизнес-правило: 5-6, 7-9, 10-11. Живёт в сервисе, а не в хендлере или БД.
func ClassGroupByNumber(classNum int) (string, error) {
	switch {
	case classNum >= 5 && classNum <= 6:
		return "5-6", nil
	case classNum >= 7 && classNum <= 9:
		return "7-9", nil
	case classNum >= 10 && classNum <= 11:
		return "10-11", nil
	default:
		return "", fmt.Errorf("неверный номер класса: %d (допустимо 5–11)", classNum)
	}
}

// UserService — сервисный слой для регистрации пользователей.
// Содержит бизнес-логику между хендлером и репозиторием:
// определение группы по номеру класса, сборка структуры User.
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService создаёт новый экземпляр UserService.
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// Register регистрирует нового пользователя.
// 1. Определяет группу классов по номеру класса.
// 2. Находит ID группы в БД.
// 3. Создаёт запись пользователя.
func (s *UserService) Register(ctx context.Context, chatID int64, username, name, phone string, classNum int) (*repository.User, error) {
	groupName, err := ClassGroupByNumber(classNum)
	if err != nil {
		return nil, err
	}

	groupID, err := s.userRepo.GetGroupIDByName(ctx, groupName)
	if err != nil {
		return nil, fmt.Errorf("группа не найдена: %w", err)
	}

	u := &repository.User{
		TgChatID:     chatID,
		TgUsername:   username,
		FullName:     name,
		Phone:        phone,
		ClassNumber:  classNum,
		ClassGroupID: groupID,
	}
	id, err := s.userRepo.Create(ctx, u)
	if err != nil {
		return nil, err
	}
	u.ID = id
	return u, nil
}
