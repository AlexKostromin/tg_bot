package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.com/AlexKostromin/tg_bot/internal/fsm"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

func (h *BookingsHandler) Handle(ctx context.Context, msg *tgbotapi.Message, user *repository.User) {
	h.ShowBookings(ctx, msg.Chat.ID, user.ID)
}

func (h *BookingsHandler) ShowBookings(ctx context.Context, chatID int64, userID int) {
	bookings, err := h.bookingRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		h.api.Send(tgbotapi.NewMessage(chatID, "Произошла ошибка. Попробуйте позже."))
		return
	}

	if len(bookings) == 0 {
		h.api.Send(tgbotapi.NewMessage(chatID, "У вас нет активных записей."))
		return
	}

	for _, b := range bookings {
		status := map[string]string{
			"pending":   "⏳ Ожидает подтверждения",
			"confirmed": "✅ Подтверждено",
		}[b.Status]

		text := fmt.Sprintf(
			"*%s*\n📅 %s, %s – %s\n%s",
			b.SubjectName,
			b.SlotDate.Format("02.01.2006"), b.StartTime, b.EndTime,
			status,
		)
		msg := tgbotapi.NewMessage(chatID, text)
		msg.ParseMode = "Markdown"

		if b.Status == "pending" || b.Status == "confirmed" {
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData(
						"🔄 Перенести", fmt.Sprintf("reschedule:%d", b.ID),
					),
					tgbotapi.NewInlineKeyboardButtonData(
						"❌ Отменить", fmt.Sprintf("cancel_booking:%d", b.ID),
					),
				),
			)
		}
		h.api.Send(msg)
	}
}

func (h *BookingsHandler) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	parts := strings.SplitN(cb.Data, ":", 2)
	if len(parts) < 2 {
		return
	}
	action, arg := parts[0], parts[1]

	switch action {
	case "cancel_booking":
		h.handleCancel(ctx, cb, arg)
	case "reschedule":
		h.handleRescheduleStart(ctx, cb, arg)
	case "reschedule_date":
		h.handleRescheduleDate(ctx, cb, arg)
	case "reschedule_slot":
		h.handleRescheduleSlot(ctx, cb, arg)
	}
}

func (h *BookingsHandler) handleCancel(ctx context.Context, cb *tgbotapi.CallbackQuery, arg string) {
	bookingID, _ := strconv.Atoi(arg)

	user, _ := h.userRepo.GetByChatID(ctx, cb.Message.Chat.ID)
	if err := h.bookingRepo.Cancel(ctx, bookingID, user.ID); err != nil {
		h.api.Request(tgbotapi.NewCallback(cb.ID, "Не удалось отменить"))
		return
	}

	edit := tgbotapi.NewEditMessageText(
		cb.Message.Chat.ID, cb.Message.MessageID,
		cb.Message.Text+"\n\n❌ *Отменено*",
	)
	edit.ParseMode = "Markdown"
	h.api.Send(edit)

	go h.notifyTutorCancellation(context.Background(), bookingID)
}

func (h *BookingsHandler) handleRescheduleStart(ctx context.Context, cb *tgbotapi.CallbackQuery, arg string) {
	chatID := cb.Message.Chat.ID
	bookingID, _ := strconv.Atoi(arg)

	info, err := h.bookingRepo.GetFullInfo(ctx, bookingID)
	if err != nil {
		h.api.Request(tgbotapi.NewCallback(cb.ID, "Запись не найдена"))
		return
	}

	user, _ := h.userRepo.GetByChatID(ctx, chatID)
	dates, err := h.slotRepo.GetAvailableDates(ctx, user.ClassGroupID, info.SubjectID)
	if err != nil || len(dates) == 0 {
		h.api.Request(tgbotapi.NewCallback(cb.ID, "Нет свободных дат"))
		return
	}

	h.fsm.Set(ctx, chatID, &fsm.UserState{
		State: fsm.StateRescheduleAwaitDate,
		Data: map[string]string{
			"booking_id": arg,
			"subject_id": strconv.Itoa(info.SubjectID),
		},
	})

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, d := range dates {
		label := fmt.Sprintf("%s (%s)", d.Format("02.01"), russianWeekday[d.Weekday()])
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("reschedule_date:%s", d.Format("2006-01-02"))),
		))
	}
	edit := tgbotapi.NewEditMessageText(chatID, cb.Message.MessageID, "Выберите новую дату:")
	edit.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	h.api.Send(edit)
}

func (h *BookingsHandler) handleRescheduleDate(ctx context.Context, cb *tgbotapi.CallbackQuery, arg string) {
	chatID := cb.Message.Chat.ID

	state, _ := h.fsm.Get(ctx, chatID)
	sid, _ := strconv.Atoi(state.Data["subject_id"])
	date, _ := time.Parse("2006-01-02", arg)

	user, _ := h.userRepo.GetByChatID(ctx, chatID)
	slots, _ := h.slotRepo.GetAvailableSlots(ctx, user.ClassGroupID, sid, date)
	if len(slots) == 0 {
		h.api.Request(tgbotapi.NewCallback(cb.ID, "Нет свободного времени"))
		return
	}

	h.fsm.TransitionWithData(ctx, chatID, fsm.StateRescheduleAwaitSlot, "date", arg)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, sl := range slots {
		label := fmt.Sprintf("%s – %s", sl.StartTime, sl.EndTime)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("reschedule_slot:%d", sl.ID)),
		))
	}
	edit := tgbotapi.NewEditMessageText(chatID, cb.Message.MessageID,
		fmt.Sprintf("Доступное время на %s:", date.Format("02.01.2006")))
	edit.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
	h.api.Send(edit)
}

func (h *BookingsHandler) handleRescheduleSlot(ctx context.Context, cb *tgbotapi.CallbackQuery, arg string) {
	chatID := cb.Message.Chat.ID

	state, _ := h.fsm.Get(ctx, chatID)
	oldBookingID, _ := strconv.Atoi(state.Data["booking_id"])
	newSlotID, _ := strconv.Atoi(arg)

	user, _ := h.userRepo.GetByChatID(ctx, chatID)

	// Отменяем старую бронь
	if err := h.bookingRepo.Cancel(ctx, oldBookingID, user.ID); err != nil {
		h.api.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось перенести: "+err.Error()))
		h.fsm.Clear(ctx, chatID)
		return
	}

	// Создаём новую
	newBooking, err := h.bookingSvc.Book(ctx, user.ID, newSlotID, fmt.Sprintf("перенос с #%d", oldBookingID))
	if err != nil {
		h.api.Send(tgbotapi.NewMessage(chatID, "❌ Не удалось перенести: "+err.Error()))
		h.fsm.Clear(ctx, chatID)
		return
	}

	h.fsm.Clear(ctx, chatID)

	slot, _ := h.slotRepo.GetByID(ctx, newSlotID)
	text := fmt.Sprintf(
		"✅ Запись перенесена!\n\nНовая дата: %s\nВремя: %s – %s\n\nЗапись #%d ожидает подтверждения.",
		slot.SlotDate.Format("02.01.2006"), slot.StartTime, slot.EndTime, newBooking.ID,
	)
	edit := tgbotapi.NewEditMessageText(chatID, cb.Message.MessageID, text)
	h.api.Send(edit)

	go h.notifyTutorReschedule(context.Background(), oldBookingID, newBooking.ID)
}

func (h *BookingsHandler) notifyTutorCancellation(ctx context.Context, bookingID int) {
	info, err := h.bookingRepo.GetFullInfo(ctx, bookingID)
	if err != nil {
		log.Error().Err(err).Msg("get booking info for cancellation failed")
		return
	}
	if info.TutorChatID == nil {
		return
	}
	text := fmt.Sprintf("❌ Запись #%d отменена учеником %s", info.BookingID, info.StudentName)
	h.api.Send(tgbotapi.NewMessage(*info.TutorChatID, text))
}

func (h *BookingsHandler) notifyTutorReschedule(ctx context.Context, oldBookingID, newBookingID int) {
	newInfo, err := h.bookingRepo.GetFullInfo(ctx, newBookingID)
	if err != nil {
		log.Error().Err(err).Msg("get booking info for reschedule failed")
		return
	}
	if newInfo.TutorChatID == nil {
		return
	}
	text := fmt.Sprintf(
		"🔄 Ученик %s перенёс запись #%d\n\nНовая запись #%d\nДата: %s\nВремя: %s – %s",
		newInfo.StudentName, oldBookingID, newBookingID,
		newInfo.SlotDate.Format("02.01.2006"), newInfo.StartTime, newInfo.EndTime,
	)
	h.api.Send(tgbotapi.NewMessage(*newInfo.TutorChatID, text))
}
