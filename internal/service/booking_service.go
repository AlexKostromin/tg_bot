package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/AlexKostromin/tg_bot/internal/config"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

// BookingService — сервисный слой для бронирования занятий.
// Содержит бизнес-правила: проверка лимита активных броней,
// проверка конфликтов по времени, создание брони.
type BookingService struct {
	bookingRepo *repository.BookingRepository
	slotRepo    *repository.SlotRepository
	cfg         *config.Config
}

// NewBookingService создаёт новый экземпляр BookingService.
func NewBookingService(bookingRepo *repository.BookingRepository, slotRepo *repository.SlotRepository, cfg *config.Config) *BookingService {
	return &BookingService{
		bookingRepo: bookingRepo,
		slotRepo:    slotRepo,
		cfg:         cfg,
	}
}

// Book создаёт бронь на занятие.
// 1. Проверяет лимит активных записей (MaxActiveBookings из конфига).
// 2. Проверяет конфликт по времени с существующими бронями.
// 3. Создаёт запись в БД через BookingRepository.Create (в транзакции).
func (s *BookingService) Book(ctx context.Context, userID, slotID int, comment string) (*repository.Booking, error) {
	// Проверяем лимит активных броней
	count, err := s.bookingRepo.CountActive(ctx, userID)
	if err != nil {
		return nil, err
	}
	if count >= s.cfg.MaxActiveBookings {
		return nil, fmt.Errorf("достигнут лимит активных записей (%d)", s.cfg.MaxActiveBookings)
	}

	// Проверяем конфликт по времени
	conflict, err := s.bookingRepo.HasConflict(ctx, userID, slotID)
	if err != nil {
		return nil, err
	}
	if conflict {
		return nil, errors.New("в это время уже есть другая запись")
	}

	b := &repository.Booking{UserID: userID, SlotID: slotID, Comment: comment}
	id, err := s.bookingRepo.Create(ctx, b)
	if err != nil {
		return nil, err
	}
	b.ID = id
	return b, nil
}
