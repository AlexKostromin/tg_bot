package handler

import (
	"github.com/username/tg_bot/internal/repository"
	"net/http"
)

type StatsHandler struct {
	repo *repository.StatsRepository
}

func NewStatsHandler(repo *repository.StatsRepository) *StatsHandler {
	return &StatsHandler{repo: repo}
}

// GetStats godoc
// @Summary      Статистика для дашборда
// @Tags         stats
// @Produce      json
// @Success      200  {object}  dto.StatsResponse
// @Router       /admin/stats [get]
// @Security     BearerAuth
func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.repo.GetDashboardStats(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, stats)
}
