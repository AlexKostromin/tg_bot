package bot

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

type NotificationService struct {
	BookingRepo *repository.BookingRepository
	api         *tgbotapi.BotAPI
}

func (n *NotificationService) SetAPI(api *tgbotapi.BotAPI) {
	n.api = api
}

func (n *NotificationService) HandleNewBooking(ctx context.Context, bookingID int) error {
	info, err := n.BookingRepo.GetFullInfo(ctx, bookingID)
	if err != nil {
		return err
	}
	if info.TutorChatID == nil {
		log.Warn().Int("booking_id", bookingID).Msg("tutor has no chat_id, skip notify")
		return nil
	}
	text := fmt.Sprintf(
		"🔔 Новая запись #%d\n\nУченик: %s\nКласс: %d\nТелефон: %s\n\nПредмет: %s\nДата: %s\nВремя: %s – %s",
		info.BookingID, info.StudentName, info.ClassNumber, info.Phone,
		info.SubjectName, info.SlotDate.Format("02.01.2006"), info.StartTime, info.EndTime,
	)
	msg := tgbotapi.NewMessage(*info.TutorChatID, text)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Подтвердить",
				fmt.Sprintf("admin_confirm:%d", bookingID)),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отклонить",
				fmt.Sprintf("admin_reject:%d", bookingID)),
		),
	)
	_, err = n.api.Send(msg)
	return err
}

func (n *NotificationService) HandleBookingStatusChanged(ctx context.Context, bookingID int, status string) error {
	info, err := n.BookingRepo.GetFullInfo(ctx, bookingID)
	if err != nil {
		return err
	}

	var statusEmoji string
	switch status {
	case "confirmed":
		statusEmoji = "✅"
	case "cancelled":
		statusEmoji = "❌"
	case "completed":
		statusEmoji = "✔️"
	default:
		statusEmoji = "ℹ️"
	}

	// Уведомить репетитора об изменении статуса
	if info.TutorChatID != nil {
		text := fmt.Sprintf(
			"%s Статус записи #%d изменился на: <b>%s</b>\n\nУченик: %s\nПредмет: %s\nДата: %s\nВремя: %s – %s",
			statusEmoji, bookingID, status, info.StudentName, info.SubjectName,
			info.SlotDate.Format("02.01.2006"), info.StartTime, info.EndTime,
		)
		msg := tgbotapi.NewMessage(*info.TutorChatID, text)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err := n.api.Send(msg); err != nil {
			log.Error().Err(err).Int("booking_id", bookingID).Msg("failed to notify tutor")
		}
	}

	return nil
}
