package handler

import (
	"bet/internal/domain"
	"bet/internal/middleware"
	"bet/internal/service"
	"bet/internal/validator"
	"encoding/json"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

type BetHandler struct {
	service   service.BetServiceUseCase
	validator validator.BetValidator
	logger    *zap.Logger
}

func NewBetHandler(service service.BetServiceUseCase, validator validator.BetValidator, logger *zap.Logger) *BetHandler {
	return &BetHandler{
		service:   service,
		validator: validator,
		logger:    logger,
	}
}

func (h *BetHandler) CreateBet(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	contentType := r.Header.Get("Content-Type")
	if idx := strings.Index(contentType, ";"); idx != -1 {
		contentType = contentType[:idx]
	}
	contentType = strings.TrimSpace(contentType)
	if contentType != "application/json" {
		h.logger.Warn("invalid content type",
			zap.String("request_id", requestID),
			zap.String("content_type", r.Header.Get("Content-Type")),
		)
		handleError(w, r, &domain.ValidationError{
			Field:   "content-type",
			Message: "content type must be application/json",
		}, h.logger)
		return
	}

	const maxBodySize = 1 << 20
	if r.ContentLength > maxBodySize {
		h.logger.Warn("request body too large",
			zap.String("request_id", requestID),
			zap.Int64("size", r.ContentLength),
		)
		handleError(w, r, &domain.ValidationError{
			Field:   "body",
			Message: "request body too large",
		}, h.logger)
		return
	}

	var req struct {
		UserID     int64   `json:"user_id"`
		Amount     float64 `json:"amount"`
		CrashPoint float64 `json:"crash_point"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		handleError(w, r, &domain.ValidationError{
			Field:   "body",
			Message: "invalid request body",
		}, h.logger)
		return
	}

	if err := h.validator.ValidateNumber(req.Amount); err != nil {
		handleError(w, r, err, h.logger)
		return
	}
	if err := h.validator.ValidateNumber(req.CrashPoint); err != nil {
		handleError(w, r, err, h.logger)
		return
	}

	if err := h.validator.ValidateCreateRequest(req.UserID, req.Amount, req.CrashPoint); err != nil {
		handleError(w, r, err, h.logger)
		return
	}

	bet, err := h.service.CreateBet(r.Context(), req.UserID, req.Amount, req.CrashPoint)
	if err != nil {
		handleError(w, r, err, h.logger)
		return
	}

	betDTO := BetDTOFromDomain(bet)
	sendJSON(w, http.StatusCreated, betDTO, h.logger)
}

func (h *BetHandler) GetBet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.validator.ValidateBetID(id); err != nil {
		handleError(w, r, err, h.logger)
		return
	}

	bet, err := h.service.GetBetByID(r.Context(), id)
	if err != nil {
		handleError(w, r, err, h.logger)
		return
	}

	betDTO := BetDTOFromDomain(bet)
	sendJSON(w, http.StatusOK, betDTO, h.logger)
}

func (h *BetHandler) ListBets(w http.ResponseWriter, r *http.Request) {
	listReq := ParseListBetsRequest(r)

	if err := h.validator.ValidatePagination(listReq.Pagination.Page, listReq.Pagination.Limit); err != nil {
		handleError(w, r, err, h.logger)
		return
	}

	if err := h.validator.ValidateSort(listReq.Sort.SortBy, listReq.Sort.Order); err != nil {
		handleError(w, r, err, h.logger)
		return
	}

	response, err := h.service.ListBets(r.Context(), listReq)
	if err != nil {
		handleError(w, r, err, h.logger)
		return
	}

	responseDTO := ListBetsResponseDTOFromDomain(response)
	sendJSON(w, http.StatusOK, responseDTO, h.logger)
}
