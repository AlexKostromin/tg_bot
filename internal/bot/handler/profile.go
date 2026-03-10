package handler

import (
	"context"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

// Handle обрабатывает сообщения в процессе редактирования профиля.
func (h *ProfileHandler) Handle(ctx context.Context, msg *tgbotapi.Message, user *repository.User) {
	// TODO: реализовать редактирование профиля
}
