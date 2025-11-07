package handler

import (
	"context"
	"net/http"
	"time"

	"bet/internal/repository"

	"go.uber.org/zap"
)

type HealthHandler struct {
	logger    *zap.Logger
	repo      repository.BetRepository
	startTime time.Time
}

func NewHealthHandler(logger *zap.Logger, repo repository.BetRepository) *HealthHandler {
	return &HealthHandler{
		logger:    logger,
		repo:      repo,
		startTime: time.Now(),
	}
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
	Version   string `json:"version"`
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(h.startTime)

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    uptime.String(),
		Version:   "1.0.0",
	}

	sendJSON(w, http.StatusOK, response, h.logger)
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err := h.repo.HealthCheck(ctx); err != nil {
		h.logger.Warn("repository health check failed", zap.Error(err))
		response := map[string]string{
			"status": "not ready",
			"error":  "repository unavailable",
		}
		sendJSON(w, http.StatusServiceUnavailable, response, h.logger)
		return
	}

	response := map[string]string{
		"status": "ready",
	}
	sendJSON(w, http.StatusOK, response, h.logger)
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "alive",
	}
	sendJSON(w, http.StatusOK, response, h.logger)
}
