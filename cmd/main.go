// @title           Tutor Bot API
// @version         1.0
// @description     API for managing tutoring bookings
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/admin
// @schemes   http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"github.com/AlexKostromin/tg_bot/docs"
	"github.com/AlexKostromin/tg_bot/internal/api"
	"github.com/AlexKostromin/tg_bot/internal/bot"
	"github.com/AlexKostromin/tg_bot/internal/config"
	"github.com/AlexKostromin/tg_bot/internal/db"
	"github.com/AlexKostromin/tg_bot/internal/fsm"
	"github.com/AlexKostromin/tg_bot/internal/mq"
	"github.com/AlexKostromin/tg_bot/internal/repository"
	"github.com/AlexKostromin/tg_bot/internal/scheduler"
	"github.com/AlexKostromin/tg_bot/internal/service"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	// Swagger
	docs.SwaggerInfo.Title = "Tutor Bot API"
	docs.SwaggerInfo.Description = "API for managing tutoring bookings"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api/admin"
	docs.SwaggerInfo.Schemes = []string{"http"}

	// PostgreSQL
	database, err := db.NewPostgres(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	if err := db.RunMigration(database); err != nil {
		log.Fatal().Err(err).Msg("migrations failed")
	}

	// Redis
	rdb := db.NewRedis(cfg)

	// RabbitMQ
	mqConn, err := mq.NewConnection(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to rabbitmq")
	}
	defer mqConn.Close()

	publisher := mq.NewPublisher(mqConn)
	_ = publisher // используется в будущих итерациях для schedule handler

	// Repositories
	userRepo := repository.NewUserRepository(database)
	slotRepo := repository.NewSlotRepository(database)
	bookingRepo := repository.NewBookingRepository(database)
	tutorRepo := repository.NewTutorRepository(database)
	subjectRepo := repository.NewSubjectRepository(database)
	statsRepo := repository.NewStatsRepository(database)
	adminUserRepo := repository.NewAdminUserRepository(database)

	// Services
	userSvc := service.NewUserService(userRepo)
	bookingSvc := service.NewBookingService(bookingRepo, slotRepo, cfg)

	// FSM
	fsmStorage := fsm.NewStorage(rdb)

	// Graceful shutdown context
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// RabbitMQ consumer
	notifySvc := &bot.NotificationService{BookingRepo: bookingRepo}
	consumer := mq.NewConsumer(mqConn, notifySvc)
	go consumer.Run(ctx)

	// Slot scheduler — поддерживает окно слотов в 2 недели для всех репетиторов
	slotScheduler := scheduler.NewSlotScheduler(database)
	go slotScheduler.Run(ctx)

	// HTTP API server
	apiServer := api.NewServer(api.Dependencies{
		Config:        cfg,
		UserRepo:      userRepo,
		SlotRepo:      slotRepo,
		BookingRepo:   bookingRepo,
		TutorRepo:     tutorRepo,
		SubjectRepo:   subjectRepo,
		StatsRepo:     statsRepo,
		AdminUserRepo: adminUserRepo,
		Publisher:     publisher,
	})
	go func() {
		if err := apiServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Telegram bot
	b, err := bot.New(cfg, bot.Dependencies{
		UserRepo:    userRepo,
		SlotRepo:    slotRepo,
		SubjectRepo: subjectRepo,
		BookingRepo: bookingRepo,
		FSM:         fsmStorage,
		UserSvc:     userSvc,
		BookingSvc:  bookingSvc,
		Publisher:   publisher,
		AdminChatID: cfg.AdminChatID,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("bot init failed")
	}
	notifySvc.SetAPI(b.API())

	go b.Run(cfg)

	<-ctx.Done()
	log.Info().Msg("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	apiServer.Shutdown(shutdownCtx)
}
