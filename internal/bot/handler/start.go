package handler

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.com/AlexKostromin/tg_bot/internal/bot/keyboard"
	"github.com/AlexKostromin/tg_bot/internal/fsm"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

func (h *StartHandler) Handle(ctx context.Context, msg *tgbotapi.Message, user *repository.User) {
	chatID := msg.Chat.ID

	if user != nil {
		reply := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("С возвращением, %s! Выберите действие:", user.FullName))
		reply.ReplyMarkup = keyboard.MainMenu()
		h.api.Send(reply)
		return
	}

	if err := h.fsm.Transition(ctx, chatID, fsm.StateRegAwaitName); err != nil {
		log.Error().Err(err).Msg("transition failed")
		return
	}
	h.api.Send(tgbotapi.NewMessage(chatID,
		"Добро пожаловать! Для начала давайте познакомимся.\n\nКак вас зовут? (Имя и фамилия)"))
}
