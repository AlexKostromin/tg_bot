package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

func (h *AdminHandler) Handle(ctx context.Context, msg *tgbotapi.Message, user *repository.User) {
	// TODO: admin slot management via text messages
	h.api.Send(tgbotapi.NewMessage(msg.Chat.ID, "Админ-панель: используйте кнопки для управления бронями."))
}

func (h *AdminHandler) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	parts := strings.SplitN(cb.Data, ":", 2)
	action := parts[0]
	bookingID, _ := strconv.Atoi(parts[1])

	switch action {
	case "admin_confirm":
		h.bookingRepo.UpdateStatus(ctx, bookingID, "confirmed")
		h.api.Send(tgbotapi.NewEditMessageText(
			cb.Message.Chat.ID, cb.Message.MessageID,
			cb.Message.Text+"\n\n✅ Подтверждено",
		))
		info, _ := h.bookingRepo.GetFullInfo(ctx, bookingID)
		h.api.Send(tgbotapi.NewMessage(info.StudentChatID,
			fmt.Sprintf("✅ Ваша запись на %s %s подтверждена!",
				info.SubjectName, info.SlotDate.Format("02.01.2006"))))

	case "admin_reject":
		h.bookingRepo.UpdateStatus(ctx, bookingID, "cancelled")
		h.api.Send(tgbotapi.NewEditMessageText(
			cb.Message.Chat.ID, cb.Message.MessageID,
			cb.Message.Text+"\n\n❌ Отклонено",
		))
		info, _ := h.bookingRepo.GetFullInfo(ctx, bookingID)
		h.api.Send(tgbotapi.NewMessage(info.StudentChatID,
			"К сожалению, ваша запись была отклонена. Попробуйте выбрать другое время."))
	}
}
