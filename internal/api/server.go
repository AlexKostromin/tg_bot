package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/log"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/AlexKostromin/tg_bot/internal/api/handler"
	"github.com/AlexKostromin/tg_bot/internal/api/middleware"
	"github.com/AlexKostromin/tg_bot/internal/config"
	"github.com/AlexKostromin/tg_bot/internal/mq"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

type Server struct {
	httpServer *http.Server
}

type Dependencies struct {
	Config        *config.Config
	UserRepo      *repository.UserRepository
	SlotRepo      *repository.SlotRepository
	BookingRepo   *repository.BookingRepository
	TutorRepo     *repository.TutorRepository
	SubjectRepo   *repository.SubjectRepository
	StatsRepo     *repository.StatsRepository
	AdminUserRepo *repository.AdminUserRepository
	Publisher     *mq.Publisher
}

func NewServer(deps Dependencies) *Server {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "https://your-domain.com"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposedHeaders:   []string{"X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Swagger
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	// API routes
	r.Route("/api/admin", func(r chi.Router) {
		// Login — без JWT
		auth := handler.NewAuthHandler(deps.AdminUserRepo, deps.Config.AdminJWTSecret)
		r.Post("/login", auth.Login)

		// Остальные роуты — под JWT-защитой
		r.Group(func(r chi.Router) {
			r.Use(middleware.JWTAuth(deps.Config.AdminJWTSecret))
			mountRoutes(r, deps)
		})
	})

	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%s", deps.Config.HTTPPort),
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	log.Info().Str("addr", s.httpServer.Addr).Msg("starting HTTP server")
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
