package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.com/AlexKostromin/tg_bot/internal/bot/keyboard"
	"github.com/AlexKostromin/tg_bot/internal/fsm"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

func (h *RegisterHandler) Handle(ctx context.Context, msg *tgbotapi.Message, user *repository.User) {
	chatID := msg.Chat.ID

	state, err := h.fsm.Get(ctx, chatID)
	if err != nil {
		return

	}

	switch state.State {
	case fsm.StateNone:
		h.fsm.Transition(ctx, chatID, fsm.StateRegAwaitName)
		h.api.Send(tgbotapi.NewMessage(chatID, "Как вас зовут? (Имя и фамилия)"))

	case fsm.StateRegAwaitName:
		name := strings.TrimSpace(msg.Text)
		if len(name) < 2 {
			h.api.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, введите полное имя."))
			return
		}
		h.fsm.TransitionWithData(ctx, chatID, fsm.StateRegAwaitPhone, "name", name)

		kb := tgbotapi.NewReplyKeyboard(
			tgbotapi.NewKeyboardButtonRow(
				tgbotapi.NewKeyboardButtonContact("📱 Отправить номер"),
			),
		)
		kb.OneTimeKeyboard = true
		reply := tgbotapi.NewMessage(chatID, "Поделитесь номером телефона:")
		reply.ReplyMarkup = kb
		h.api.Send(reply)

	case fsm.StateRegAwaitPhone:
		var phone string
		if msg.Contact != nil {
			phone = msg.Contact.PhoneNumber
		} else {
			phone = strings.TrimSpace(msg.Text)
		}
		if phone == "" {
			h.api.Send(tgbotapi.NewMessage(chatID, "Пожалуйста, поделитесь номером телефона."))
			return
		}
		h.fsm.TransitionWithData(ctx, chatID, fsm.StateRegAwaitClass, "phone", phone)

		reply := tgbotapi.NewMessage(chatID, "В каком классе вы учитесь?")
		reply.ReplyMarkup = keyboard.ClassNumbers()
		h.api.Send(reply)

	case fsm.StateRegAwaitClass:
		classNum, err := strconv.Atoi(strings.TrimSpace(msg.Text))
		if err != nil || classNum < 5 || classNum > 11 {
			h.api.Send(tgbotapi.NewMessage(chatID, "Выберите класс из предложенных вариантов (5–11)."))
			return
		}

		_, err = h.userSvc.Register(ctx, chatID,
			msg.From.UserName,
			state.Data["name"],
			state.Data["phone"],
			classNum,
		)
		if err != nil {
			log.Error().Err(err).Msg("register failed")
			h.api.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка. Попробуйте позже."))
			return
		}

		h.fsm.Clear(ctx, chatID)
		reply := tgbotapi.NewMessage(chatID,
			fmt.Sprintf("✅ Регистрация завершена! Добро пожаловать, %s!", state.Data["name"]))
		reply.ReplyMarkup = keyboard.MainMenu()
		h.api.Send(reply)
	}
}
