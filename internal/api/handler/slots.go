// internal/api/handler/slots.go

package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/AlexKostromin/tg_bot/internal/api/dto"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

type SlotHandler struct {
	repo *repository.SlotRepository
}

func NewSlotHandler(repo *repository.SlotRepository) *SlotHandler {
	return &SlotHandler{repo: repo}
}

// List godoc
// @Summary      Список временных слотов
// @Tags         slots
// @Produce      json
// @Param        date        query   string  false  "Фильтр по дате (YYYY-MM-DD)"
// @Param        group_id    query   int     false  "ID группы классов"
// @Param        available   query   bool    false  "Только доступные"
// @Param        page        query   int     false  "Страница" default(1)
// @Param        limit       query   int     false  "Размер страницы" default(20)
// @Success      200  {object}  dto.SlotListResponse
// @Router       /admin/slots [get]
// @Security     BearerAuth
func (h *SlotHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	pag := dto.PaginatedRequest{
		Page:  queryInt(r, "page", 1),
		Limit: queryInt(r, "limit", 20),
	}
	pag.Defaults()

	filters := repository.SlotFilters{
		Date:      r.URL.Query().Get("date"),
		GroupID:   queryInt(r, "group_id", 0),
		Available: r.URL.Query().Get("available"),
		Offset:    pag.Offset(),
		Limit:     pag.Limit,
	}

	items, total, err := h.repo.ListWithFilters(ctx, filters)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto.SlotListResponse{
		Items: items,
		Total: total,
		Page:  pag.Page,
		Limit: pag.Limit,
	})
}

// Create godoc
// @Summary      Создать слот
// @Tags         slots
// @Accept       json
// @Produce      json
// @Param        body  body  dto.CreateSlotRequest  true  "Данные слота"
// @Success      201   {object}  dto.SlotResponse
// @Failure      422   {object}  dto.ValidationErrorResponse
// @Router       /admin/slots [post]
// @Security     BearerAuth
func (h *SlotHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	slot, err := h.repo.Create(r.Context(), repository.CreateSlotParams{
		TutorID:      req.TutorID,
		SubjectID:    req.SubjectID,
		ClassGroupID: req.ClassGroupID,
		SlotDate:     req.SlotDate,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, slot)
}

// BulkCreate godoc
// @Summary      Создать несколько слотов
// @Tags         slots
// @Accept       json
// @Produce      json
// @Param        body  body  dto.BulkCreateSlotRequest  true  "Массив слотов"
// @Success      201   {object}  dto.SlotListResponse
// @Router       /admin/slots/bulk [post]
// @Security     BearerAuth
func (h *SlotHandler) BulkCreate(w http.ResponseWriter, r *http.Request) {
	var req dto.BulkCreateSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var created []*repository.TimeSlot
	for _, s := range req.Slots {
		slot, err := h.repo.Create(r.Context(), repository.CreateSlotParams{
			TutorID:      s.TutorID,
			SubjectID:    s.SubjectID,
			ClassGroupID: s.ClassGroupID,
			SlotDate:     s.SlotDate,
			StartTime:    s.StartTime,
			EndTime:      s.EndTime,
		})
		if err != nil {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		created = append(created, slot)
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"items": created,
		"total": len(created),
	})
}

// GetByID godoc
// @Summary      Один слот
// @Tags         slots
// @Produce      json
// @Param        id  path  int  true  "ID слота"
// @Success      200  {object}  dto.SlotResponse
// @Router       /admin/slots/{id} [get]
// @Security     BearerAuth
func (h *SlotHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	slot, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "slot not found")
		return
	}
	writeJSON(w, http.StatusOK, slot)
}

// Update godoc
// @Summary      Обновить слот
// @Tags         slots
// @Accept       json
// @Produce      json
// @Param        id    path  int                    true  "ID слота"
// @Param        body  body  dto.CreateSlotRequest  true  "Данные слота"
// @Success      200   {object}  dto.SlotResponse
// @Router       /admin/slots/{id} [put]
// @Security     BearerAuth
func (h *SlotHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req dto.CreateSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	slot, err := h.repo.Update(r.Context(), id, repository.CreateSlotParams{
		TutorID:      req.TutorID,
		SubjectID:    req.SubjectID,
		ClassGroupID: req.ClassGroupID,
		SlotDate:     req.SlotDate,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, slot)
}

// Delete godoc
// @Summary      Удалить слот
// @Tags         slots
// @Param        id  path  int  true  "ID слота"
// @Success      204
// @Failure      409  {object}  dto.ErrorResponse  "Слот забронирован"
// @Router       /admin/slots/{id} [delete]
// @Security     BearerAuth
func (h *SlotHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
