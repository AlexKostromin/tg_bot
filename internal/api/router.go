package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/username/tg_bot/internal/api/handler"
)

func mountRoutes(r chi.Router, deps Dependencies) {
	stats := handler.NewStatsHandler(deps.StatsRepo)
	r.Get("/stats", stats.GetStats)

	users := handler.NewUserHandler(deps.UserRepo)
	r.Get("/users", users.List)
	r.Get("/users/{id}", users.GetByID)
	r.Patch("/users/{id}", users.Update)

	slots := handler.NewSlotHandler(deps.SlotRepo)
	r.Get("/slots", slots.List)
	r.Post("/slots", slots.Create)
	r.Post("/slots/bulk", slots.BulkCreate)
	r.Get("/slots/{id}", slots.GetByID)
	r.Put("/slots/{id}", slots.Update)
	r.Delete("/slots/{id}", slots.Delete)

	bookings := handler.NewBookingHandler(deps.BookingRepo, deps.Publisher)
	r.Get("/bookings", bookings.List)
	r.Get("/bookings/{id}", bookings.GetByID)
	r.Patch("/bookings/{id}/status", bookings.UpdateStatus)

	tutors := handler.NewTutorHandler(deps.TutorRepo)
	r.Get("/tutors", tutors.List)
	r.Post("/tutors", tutors.Create)
	r.Get("/tutors/{id}", tutors.GetByID)
	r.Put("/tutors/{id}", tutors.Update)
	r.Delete("/tutors/{id}", tutors.Delete)

	subjects := handler.NewSubjectHandler(deps.SubjectRepo)
	r.Get("/subjects", subjects.List)
	r.Post("/subjects", subjects.Create)
	r.Put("/subjects/{id}", subjects.Update)
	r.Delete("/subjects/{id}", subjects.Delete)

	r.Get("/class-groups", handler.NewClassGroupHandler(deps.SubjectRepo).List)
}
