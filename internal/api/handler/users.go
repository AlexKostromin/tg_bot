// internal/api/handler/users.go

package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/AlexKostromin/tg_bot/internal/api/dto"
	"github.com/AlexKostromin/tg_bot/internal/repository"
)

type UserHandler struct {
	repo *repository.UserRepository
}

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

// List godoc
// @Summary      Список пользователей
// @Tags         users
// @Produce      json
// @Param        search  query  string  false  "Поиск по имени/телефону"
// @Param        page    query  int     false  "Страница" default(1)
// @Param        limit   query  int     false  "Размер страницы" default(20)
// @Success      200  {object}  dto.UserListResponse
// @Router       /admin/users [get]
// @Security     BearerAuth
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	pag := dto.PaginatedRequest{
		Page:  queryInt(r, "page", 1),
		Limit: queryInt(r, "limit", 20),
	}
	pag.Defaults()

	search := r.URL.Query().Get("search")
	items, total, err := h.repo.ListWithSearch(r.Context(), search, pag.Offset(), pag.Limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, dto.UserListResponse{
		Items: items,
		Total: total,
		Page:  pag.Page,
		Limit: pag.Limit,
	})
}

// GetByID godoc
// @Summary      Один пользователь
// @Tags         users
// @Produce      json
// @Param        id  path  int  true  "ID пользователя"
// @Success      200  {object}  dto.UserResponse
// @Router       /admin/users/{id} [get]
// @Security     BearerAuth
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	user, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// Update godoc
// @Summary      Обновить пользователя (блокировка)
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path  int                    true  "ID пользователя"
// @Param        body  body  dto.UpdateUserRequest  true  "Данные"
// @Success      200   {object}  dto.UserResponse
// @Router       /admin/users/{id} [patch]
// @Security     BearerAuth
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.repo.SetActive(r.Context(), id, req.IsActive); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, _ := h.repo.GetByID(r.Context(), id)
	writeJSON(w, http.StatusOK, user)
}
