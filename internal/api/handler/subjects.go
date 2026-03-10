package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/username/tg_bot/internal/repository"
)

type SubjectHandler struct {
	repo *repository.SubjectRepository
}

func NewSubjectHandler(repo *repository.SubjectRepository) *SubjectHandler {
	return &SubjectHandler{repo: repo}
}

// List godoc
// @Summary      Список предметов
// @Tags         subjects
// @Produce      json
// @Success      200  {array}   object
// @Router       /admin/subjects [get]
// @Security     BearerAuth
func (h *SubjectHandler) List(w http.ResponseWriter, r *http.Request) {
	subjects, err := h.repo.GetAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, subjects)
}

// Create godoc
// @Summary      Создать предмет
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Param        body  body  object  true  "Данные предмета (name обязателен)"
// @Success      201   {object}  object
// @Router       /admin/subjects [post]
// @Security     BearerAuth
func (h *SubjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	subject, err := h.repo.Create(r.Context(), req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, subject)
}

// Update godoc
// @Summary      Обновить предмет
// @Tags         subjects
// @Accept       json
// @Produce      json
// @Param        id    path  int     true  "ID предмета"
// @Param        body  body  object  true  "Данные предмета"
// @Success      200   {object}  object
// @Router       /admin/subjects/{id} [put]
// @Security     BearerAuth
func (h *SubjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	subject, err := h.repo.Update(r.Context(), id, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, subject)
}

// Delete godoc
// @Summary      Удалить предмет
// @Tags         subjects
// @Param        id   path   int  true  "ID предмета"
// @Success      204
// @Router       /admin/subjects/{id} [delete]
// @Security     BearerAuth
func (h *SubjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

// --- ClassGroup handler ---

type ClassGroupHandler struct {
	repo *repository.SubjectRepository
}

func NewClassGroupHandler(repo *repository.SubjectRepository) *ClassGroupHandler {
	return &ClassGroupHandler{repo: repo}
}

// List godoc
// @Summary      Список групп классов
// @Tags         class-groups
// @Produce      json
// @Success      200  {array}   object
// @Router       /admin/class-groups [get]
// @Security     BearerAuth
func (h *ClassGroupHandler) List(w http.ResponseWriter, r *http.Request) {
	groups, err := h.repo.GetAllGroups(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, groups)
}
