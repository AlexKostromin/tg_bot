package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/AlexKostromin/tg_bot/internal/fsm"
	"github.com/AlexKostromin/tg_bot/internal/repository"
	"github.com/AlexKostromin/tg_bot/internal/service"
)

type StartHandler struct {
	api *tgbotapi.BotAPI
	fsm *fsm.Storage
}

func NewStartHandler(api *tgbotapi.BotAPI, f *fsm.Storage) *StartHandler {
	return &StartHandler{api: api, fsm: f}
}

type RegisterHandler struct {
	api     *tgbotapi.BotAPI
	fsm     *fsm.Storage
	userSvc *service.UserService
}

func NewRegisterHandler(api *tgbotapi.BotAPI, f *fsm.Storage, userSvc *service.UserService) *RegisterHandler {
	return &RegisterHandler{api: api, fsm: f, userSvc: userSvc}
}

type ScheduleHandler struct {
	api         *tgbotapi.BotAPI
	fsm         *fsm.Storage
	userRepo    *repository.UserRepository
	slotRepo    *repository.SlotRepository
	subjectRepo *repository.SubjectRepository
	bookingRepo *repository.BookingRepository
	bookingSvc  *service.BookingService
}

func NewScheduleHandler(api *tgbotapi.BotAPI, f *fsm.Storage, userRepo *repository.UserRepository, slotRepo *repository.SlotRepository, subjectRepo *repository.SubjectRepository, bookingRepo *repository.BookingRepository, bookingSvc *service.BookingService) *ScheduleHandler {
	return &ScheduleHandler{api: api, fsm: f, userRepo: userRepo, slotRepo: slotRepo, subjectRepo: subjectRepo, bookingRepo: bookingRepo, bookingSvc: bookingSvc}
}

type BookingsHandler struct {
	api         *tgbotapi.BotAPI
	fsm         *fsm.Storage
	bookingRepo *repository.BookingRepository
	slotRepo    *repository.SlotRepository
	userRepo    *repository.UserRepository
	bookingSvc  *service.BookingService
}

func NewBookingsHandler(api *tgbotapi.BotAPI, f *fsm.Storage, bookingRepo *repository.BookingRepository, slotRepo *repository.SlotRepository, userRepo *repository.UserRepository, bookingSvc *service.BookingService) *BookingsHandler {
	return &BookingsHandler{api: api, fsm: f, bookingRepo: bookingRepo, slotRepo: slotRepo, userRepo: userRepo, bookingSvc: bookingSvc}
}

type ProfileHandler struct {
	api     *tgbotapi.BotAPI
	fsm     *fsm.Storage
	userSvc *service.UserService
}

func NewProfileHandler(api *tgbotapi.BotAPI, f *fsm.Storage, userSvc *service.UserService) *ProfileHandler {
	return &ProfileHandler{api: api, fsm: f, userSvc: userSvc}
}

type AdminHandler struct {
	api         *tgbotapi.BotAPI
	fsm         *fsm.Storage
	slotRepo    *repository.SlotRepository
	bookingRepo *repository.BookingRepository
}

func NewAdminHandler(api *tgbotapi.BotAPI, f *fsm.Storage, slotRepo *repository.SlotRepository, bookingRepo *repository.BookingRepository) *AdminHandler {
	return &AdminHandler{api: api, fsm: f, slotRepo: slotRepo, bookingRepo: bookingRepo}
}
