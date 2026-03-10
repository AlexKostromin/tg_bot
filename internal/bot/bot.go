package bot

import (
	"context"
	"fmt"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.com/AlexKostromin/tg_bot/internal/bot/handler"
	"github.com/AlexKostromin/tg_bot/internal/config"
)

type Bot struct {
	api    *tgbotapi.BotAPI
	router *Router
}

func New(cfg *config.Config, deps Dependencies) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, fmt.Errorf("bot api init: %w", err)
	}
	api.Debug = cfg.LogLevel == "debug"
	log.Info().Str("username", api.Self.UserName).Msg("bot authorized")

	// Создаём хендлеры — им нужен api, который доступен только здесь
	handlers := routerHandlers{
		start:    handler.NewStartHandler(api, deps.FSM),
		register: handler.NewRegisterHandler(api, deps.FSM, deps.UserSvc),
		schedule: handler.NewScheduleHandler(api, deps.FSM, deps.UserRepo, deps.SlotRepo, deps.SubjectRepo, deps.BookingRepo, deps.BookingSvc),
		bookings: handler.NewBookingsHandler(api, deps.FSM, deps.BookingRepo, deps.SlotRepo, deps.UserRepo, deps.BookingSvc),
		profile:  handler.NewProfileHandler(api, deps.FSM, deps.UserSvc),
		admin:    handler.NewAdminHandler(api, deps.FSM, deps.SlotRepo, deps.BookingRepo),
	}

	router := NewRouter(api, deps, handlers)
	return &Bot{api: api, router: router}, nil
}

// API возвращает tgbotapi.BotAPI для использования в NotificationService.
func (b *Bot) API() *tgbotapi.BotAPI {
	return b.api
}

func (b *Bot) Run(cfg *config.Config) {
	if cfg.WebhookURL != "" {
		b.runWebhook(cfg)
	} else {
		b.runPolling()
	}
}

func (b *Bot) runPolling() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)
	for update := range updates {
		go b.router.Handle(context.Background(), update)
	}
}

func (b *Bot) runWebhook(cfg *config.Config) {
	wh, _ := tgbotapi.NewWebhook(cfg.WebhookURL + "/bot" + cfg.BotToken)
	b.api.Request(wh)

	updates := b.api.ListenForWebhook("/bot" + cfg.BotToken)
	go http.ListenAndServe(fmt.Sprintf(":%d", cfg.WebhookPort), nil)

	for update := range updates {
		go b.router.Handle(context.Background(), update)
	}
}
