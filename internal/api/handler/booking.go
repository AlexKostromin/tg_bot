package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/AlexKostromin/tg_bot/internal/api/dto"
	"github.com/AlexKostromin/tg_bot/internal/mq"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

type BookingHandler struct {
	repo      *repository.BookingRepository
	publisher *mq.Publisher
}

func NewBookingHandler(repo *repository.BookingRepository, publisher *mq.Publisher) *BookingHandler {
	return &BookingHandler{repo: repo, publisher: publisher}
}

// List godoc
// @Summary      Список броней
// @Tags         bookings
// @Produce      json
// @Param        status   query   string  false  "Фильтр по статусу"
// @Param        user_id  query   int     false  "Фильтр по ученику"
// @Param        date     query   string  false  "Фильтр по дате (YYYY-MM-DD)"
// @Param        page     query   int     false  "Страница" default(1)
// @Param        limit    query   int     false  "Размер страницы" default(20)
// @Success      200  {object}  dto.BookingListResponse
// @Router       /admin/bookings [get]
// @Security     BearerAuth
func (h *BookingHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pag := dto.PaginatedRequest{
		Page:  queryInt(r, "page", 1),
		Limit: queryInt(r, "limit", 20),
	}
	pag.Defaults()

	filters := repository.BookingFilters{
		Status: r.URL.Query().Get("status"),
		UserID: queryInt(r, "user_id", 0),
		Date:   r.URL.Query().Get("date"),
		Offset: pag.Offset(),
		Limit:  pag.Limit,
	}

	items, total, err := h.repo.ListWithFilters(ctx, filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Маппируем FullInfo в BookingResponse
	responses := make([]dto.BookingResponse, len(items))
	for i, item := range items {
		responses[i] = dto.BookingResponse{
			ID:          item.BookingID,
			UserID:      item.UserID,
			UserName:    item.StudentName,
			UserPhone:   item.Phone,
			SlotID:      item.SlotID,
			SlotDate:    item.SlotDate.Format("2006-01-02"),
			StartTime:   item.StartTime,
			EndTime:     item.EndTime,
			SubjectName: item.SubjectName,
			Status:      item.Status,
			BookedAt:    item.BookedAt.Format("2006-01-02 15:04:05"),
		}
	}

	writeJSON(w, http.StatusOK, dto.BookingListResponse{
		Items: responses,
		Total: total,
		Page:  pag.Page,
		Limit: pag.Limit,
	})
}

// GetByID godoc
// @Summary      Получить бронь по ID
// @Tags         bookings
// @Produce      json
// @Param        id   path   int  true  "ID брони"
// @Success      200  {object}  dto.BookingResponse
// @Router       /admin/bookings/{id} [get]
// @Security     BearerAuth
func (h *BookingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	booking, err := h.repo.GetFullInfo(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "booking not found")
		return
	}

	// Маппируем FullInfo в BookingResponse
	response := dto.BookingResponse{
		ID:          booking.BookingID,
		UserID:      booking.UserID,
		UserName:    booking.StudentName,
		UserPhone:   booking.Phone,
		SlotID:      booking.SlotID,
		SlotDate:    booking.SlotDate.Format("2006-01-02"),
		StartTime:   booking.StartTime,
		EndTime:     booking.EndTime,
		SubjectName: booking.SubjectName,
		Status:      booking.Status,
		BookedAt:    booking.BookedAt.Format("2006-01-02 15:04:05"),
	}

	writeJSON(w, http.StatusOK, response)
}

// UpdateStatus godoc
// @Summary      Сменить статус брони
// @Tags         bookings
// @Accept       json
// @Produce      json
// @Param        id    path  int                       true  "ID брони"
// @Param        body  body  dto.UpdateBookingStatus    true  "Новый статус"
// @Success      200   {object}  dto.BookingResponse
// @Router       /admin/bookings/{id}/status [patch]
// @Security     BearerAuth
func (h *BookingHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req dto.UpdateBookingStatus
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.repo.UpdateStatus(r.Context(), id, req.Status); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Отправить событие об изменении статуса в RabbitMQ
	if h.publisher != nil {
		h.publisher.PublishBookingStatusChanged(r.Context(), id, req.Status)
	}

	booking, _ := h.repo.GetFullInfo(r.Context(), id)
	response := dto.BookingResponse{
		ID:          booking.BookingID,
		UserID:      booking.UserID,
		UserName:    booking.StudentName,
		UserPhone:   booking.Phone,
		SlotID:      booking.SlotID,
		SlotDate:    booking.SlotDate.Format("2006-01-02"),
		StartTime:   booking.StartTime,
		EndTime:     booking.EndTime,
		SubjectName: booking.SubjectName,
		Status:      booking.Status,
		BookedAt:    booking.BookedAt.Format("2006-01-02 15:04:05"),
	}
	writeJSON(w, http.StatusOK, response)
}

// --- helpers ---

func queryInt(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, dto.ErrorResponse{Error: msg})
}
