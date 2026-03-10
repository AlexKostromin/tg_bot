package bot

import (
	"context"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.com/AlexKostromin/tg_bot/internal/bot/handler"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

type routerHandlers struct {
	start    *handler.StartHandler
	register *handler.RegisterHandler
	schedule *handler.ScheduleHandler
	bookings *handler.BookingsHandler
	profile  *handler.ProfileHandler
	admin    *handler.AdminHandler
}

type Router struct {
	api      *tgbotapi.BotAPI
	deps     Dependencies
	handlers routerHandlers
}

func NewRouter(api *tgbotapi.BotAPI, deps Dependencies, h routerHandlers) *Router {
	return &Router{api: api, deps: deps, handlers: h}
}

func (r *Router) Handle(ctx context.Context, update tgbotapi.Update) {
	switch {
	case update.Message != nil:
		r.handleMessage(ctx, update.Message)
	case update.CallbackQuery != nil:
		r.handleCallback(ctx, update.CallbackQuery)
	}
}

func (r *Router) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.Chat.ID

	user, err := r.deps.UserRepo.GetByChatID(ctx, chatID)
	if err != nil {
		log.Error().Err(err).Int64("chat_id", chatID).Msg("get user failed")
		return
	}

	if msg.IsCommand() {
		switch msg.Command() {
		case "start":
			r.handlers.start.Handle(ctx, msg, user)
		case "cancel":
			r.deps.FSM.Clear(ctx, chatID)
			r.sendText(chatID, "Действие отменено.")
		}
		return
	}

	if user == nil {
		r.handlers.register.Handle(ctx, msg, nil)
		return
	}

	state, err := r.deps.FSM.Get(ctx, chatID)
	if err != nil {
		log.Error().Err(err).Msg("get state failed")
		return
	}

	// Кнопки главного меню — всегда обрабатываем напрямую, сбрасывая FSM
	if msg.Text == "📅 Записаться" || msg.Text == "📋 Мои записи" || msg.Text == "👤 Профиль" {
		r.deps.FSM.Clear(ctx, chatID)
		r.handleMainMenu(ctx, msg, user)
		return
	}

	switch {
	case strings.HasPrefix(state.State, "reg:"):
		r.handlers.register.Handle(ctx, msg, user)
	case strings.HasPrefix(state.State, "book:"):
		r.handlers.schedule.Handle(ctx, msg, user)
	case strings.HasPrefix(state.State, "cancel:"):
		r.handlers.bookings.Handle(ctx, msg, user)
	case strings.HasPrefix(state.State, "profile:"):
		r.handlers.profile.Handle(ctx, msg, user)
	case strings.HasPrefix(state.State, "admin:"):
		if r.isAdmin(chatID) {
			r.handlers.admin.Handle(ctx, msg, user)
		}
	default:
		r.handleMainMenu(ctx, msg, user)
	}
}

func (r *Router) handleCallback(ctx context.Context, cb *tgbotapi.CallbackQuery) {
	r.api.Request(tgbotapi.NewCallback(cb.ID, ""))

	parts := strings.SplitN(cb.Data, ":", 2)
	if len(parts) < 2 {
		log.Warn().Str("data", cb.Data).Msg("invalid callback data format")
		return
	}
	action := parts[0]

	switch action {
	case "book_subject", "book_date", "book_slot":
		r.handlers.schedule.HandleCallback(ctx, cb)
	case "cancel_booking", "reschedule", "reschedule_date", "reschedule_slot":
		r.handlers.bookings.HandleCallback(ctx, cb)
	case "admin_confirm", "admin_reject":
		if r.isAdmin(cb.Message.Chat.ID) {
			r.handlers.admin.HandleCallback(ctx, cb)
		}
	}
}

func (r *Router) sendText(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	r.api.Send(msg)
}

func (r *Router) isAdmin(chatID int64) bool {
	return chatID == r.deps.AdminChatID
}

func (r *Router) handleMainMenu(ctx context.Context, msg *tgbotapi.Message, user *repository.User) {
	switch msg.Text {
	case "📅 Записаться":
		r.handlers.schedule.Handle(ctx, msg, user)
	case "📋 Мои записи":
		r.handlers.bookings.ShowBookings(ctx, msg.Chat.ID, user.ID)
	case "👤 Профиль":
		r.handlers.profile.Handle(ctx, msg, user)
	default:
		r.sendText(msg.Chat.ID, "Используйте меню для навигации.")
	}
}
