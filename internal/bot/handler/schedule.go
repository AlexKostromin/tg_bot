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

// internal/bot/handler/schedule.go

// russianWeekday возвращает сокращённое название дня недели на русском.
// time.Format("Mon") возвращает только английские названия.
var russianWeekday = map[time.Weekday]string{
	time.Monday: "Пн", time.Tuesday: "Вт", time.Wednesday: "Ср",
	time.Thursday: "Чт", time.Friday: "Пт", time.Saturday: "Сб", time.Sunday: "Вс",
}

func (h *ScheduleHandler) Handle(ctx context.Context, msg *tgbotapi.Message, user *repository.User) {
	// Вызывается из handleMainMenu при нажатии "📅 Записаться"
	h.showSubjects(ctx, msg.Chat.ID, user)
}

func (h *ScheduleHandler) showSubjects(ctx context.Context, chatID int64, user *repository.User) {
	subjects, err := h.subjectRepo.GetByGroupID(ctx, user.ClassGroupID)
	if err != nil || len(subjects) == 0 {
		h.api.Send(tgbotapi.NewMessage(chatID, "Нет доступных предметов для вашей группы."))
		return
	}

	h.fsm.Transition(ctx, chatID, fsm.StateBookAwaitSubject)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, s := range subjects {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(s.Name, fmt.Sprintf("book_subject:%d", s.ID)),
		))
	}
	reply := tgbotapi.NewMessage(chatID, "Выберите предмет:")
	reply.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	h.api.Send(reply)
}

func (h *ScheduleHandler) HandleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	parts := strings.SplitN(cb.Data, ":", 2)
	if len(parts) < 2 {
		return
	}
	action, arg := parts[0], parts[1]

	state, _ := h.fsm.Get(ctx, chatID)
	user, _ := h.userRepo.GetByChatID(ctx, chatID)

	switch action {
	case "book_subject":
		h.fsm.TransitionWithData(ctx, chatID, fsm.StateBookAwaitDate, "subject_id", arg)

		sid, _ := strconv.Atoi(arg)
		dates, err := h.slotRepo.GetAvailableDates(ctx, user.ClassGroupID, sid)
		if err != nil || len(dates) == 0 {
			h.api.Send(tgbotapi.NewMessage(chatID, "Свободных дат нет. Попробуйте позже."))
			h.fsm.Clear(ctx, chatID)
			return
		}

		var rows [][]tgbotapi.InlineKeyboardButton
		for _, d := range dates {
			label := fmt.Sprintf("%s (%s)", d.Format("02.01"), russianWeekday[d.Weekday()])
			data := fmt.Sprintf("book_date:%s", d.Format("2006-01-02"))
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(label, data),
			))
		}
		edit := tgbotapi.NewEditMessageText(chatID, cb.Message.MessageID, "Выберите дату:")
		edit.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
		h.api.Send(edit)

	case "book_date":
		h.fsm.TransitionWithData(ctx, chatID, fsm.StateBookAwaitSlot, "date", arg)

		sid, _ := strconv.Atoi(state.Data["subject_id"])
		date, _ := time.Parse("2006-01-02", arg)
		slots, _ := h.slotRepo.GetAvailableSlots(ctx, user.ClassGroupID, sid, date)

		var rows [][]tgbotapi.InlineKeyboardButton
		for _, sl := range slots {
			label := fmt.Sprintf("%s – %s", sl.StartTime, sl.EndTime)
			rows = append(rows, tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(label, fmt.Sprintf("book_slot:%d", sl.ID)),
			))
		}
		edit := tgbotapi.NewEditMessageText(chatID, cb.Message.MessageID,
			fmt.Sprintf("Доступное время на %s:", date.Format("02.01.2006")))
		edit.ReplyMarkup = &tgbotapi.InlineKeyboardMarkup{InlineKeyboard: rows}
		h.api.Send(edit)

	case "book_slot":
		slotID, _ := strconv.Atoi(arg)
		booking, err := h.bookingSvc.Book(ctx, user.ID, slotID, "")
		if err != nil {
			h.api.Send(tgbotapi.NewMessage(chatID, "❌ "+err.Error()))
			h.fsm.Clear(ctx, chatID)
			return
		}

		h.fsm.Clear(ctx, chatID)

		slot, _ := h.slotRepo.GetByID(ctx, slotID)
		text := fmt.Sprintf(
			"✅ Запись создана!\n\nДата: %s\nВремя: %s – %s\n\nЗапись #%d ожидает подтверждения.",
			slot.SlotDate.Format("02.01.2006"), slot.StartTime, slot.EndTime, booking.ID,
		)
		h.api.Send(tgbotapi.NewMessage(chatID, text))

		// Уведомляем репетитора (пока через горутину, в следующей итерации — через очередь)
		go h.notifyTutor(context.Background(), booking.ID)
	}
}

// notifyTutor отправляет уведомление репетитору о новой брони.
// В итерации 10 будет заменено на публикацию в RabbitMQ.
func (h *ScheduleHandler) notifyTutor(ctx context.Context, bookingID int) {
	info, err := h.bookingRepo.GetFullInfo(ctx, bookingID)
	if err != nil {
		log.Error().Err(err).Int("booking_id", bookingID).Msg("get booking info failed")
		return
	}
	if info.TutorChatID == nil {
		log.Warn().Int("booking_id", bookingID).Msg("tutor has no chat_id, skip notify")
		return
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
	h.api.Send(msg)
}
