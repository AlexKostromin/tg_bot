package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/username/tg_bot/internal/repository"
)

type TutorHandler struct {
	repo *repository.TutorRepository
}

func NewTutorHandler(repo *repository.TutorRepository) *TutorHandler {
	return &TutorHandler{repo: repo}
}

// List godoc
// @Summary      Список репетиторов
// @Tags         tutors
// @Produce      json
// @Success      200  {array}   object
// @Router       /admin/tutors [get]
// @Security     BearerAuth
func (h *TutorHandler) List(w http.ResponseWriter, r *http.Request) {
	tutors, err := h.repo.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tutors)
}

// Create godoc
// @Summary      Создать репетитора
// @Tags         tutors
// @Accept       json
// @Produce      json
// @Param        body  body  object  true  "Данные репетитора (full_name обязателен, tg_chat_id опционален)"
// @Success      201   {object}  object
// @Router       /admin/tutors [post]
// @Security     BearerAuth
func (h *TutorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FullName string `json:"full_name"`
		TgChatID *int64 `json:"tg_chat_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tutor, err := h.repo.Create(r.Context(), req.FullName, req.TgChatID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, tutor)
}

// GetByID godoc
// @Summary      Получить репетитора по ID
// @Tags         tutors
// @Produce      json
// @Param        id   path   int  true  "ID репетитора"
// @Success      200  {object}  object
// @Router       /admin/tutors/{id} [get]
// @Security     BearerAuth
func (h *TutorHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	tutor, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "tutor not found")
		return
	}
	writeJSON(w, http.StatusOK, tutor)
}

// Update godoc
// @Summary      Обновить репетитора
// @Tags         tutors
// @Accept       json
// @Produce      json
// @Param        id    path  int     true  "ID репетитора"
// @Param        body  body  object  true  "Данные репетитора"
// @Success      200   {object}  object
// @Router       /admin/tutors/{id} [put]
// @Security     BearerAuth
func (h *TutorHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		FullName string `json:"full_name"`
		TgChatID *int64 `json:"tg_chat_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	tutor, err := h.repo.Update(r.Context(), id, req.FullName, req.TgChatID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tutor)
}

// Delete godoc
// @Summary      Удалить репетитора
// @Tags         tutors
// @Param        id   path   int  true  "ID репетитора"
// @Success      204
// @Router       /admin/tutors/{id} [delete]
// @Security     BearerAuth
func (h *TutorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.repo.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
