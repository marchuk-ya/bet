package handler

import (
	"bet/internal/domain"
	"bet/internal/middleware"
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"
)

type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

func handleError(w http.ResponseWriter, r *http.Request, err error, logger *zap.Logger) {
	requestID := middleware.GetRequestID(r.Context())

	if r.Context().Err() != nil {
		logger.Warn("request cancelled or timeout",
			zap.String("request_id", requestID),
			zap.Error(r.Context().Err()),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		sendErrorResponse(w, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "Request cancelled or timeout", logger)
		return
	}

	var statusCode int
	var errorCode string
	var message string

	logFields := []zap.Field{
		zap.String("request_id", requestID),
		zap.Error(err),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	}

	switch {
	case domain.IsValidationError(err):
		var validationErr *domain.ValidationError
		errors.As(err, &validationErr)
		statusCode = http.StatusBadRequest
		errorCode = "VALIDATION_ERROR"
		message = validationErr.Error()
		logger.Warn("validation error", append(logFields, zap.String("error_code", errorCode))...)

	case domain.IsNotFoundError(err):
		var notFoundErr *domain.NotFoundError
		errors.As(err, &notFoundErr)
		statusCode = http.StatusNotFound
		errorCode = "BET_NOT_FOUND"
		message = notFoundErr.Error()
		logger.Info("resource not found", append(logFields, zap.String("error_code", errorCode))...)

	case domain.IsInvalidInputError(err):
		var invalidInputErr *domain.InvalidInputError
		errors.As(err, &invalidInputErr)
		statusCode = http.StatusBadRequest
		errorCode = "INVALID_INPUT"
		message = invalidInputErr.Error()
		logger.Warn("invalid input", append(logFields, zap.String("error_code", errorCode))...)

	case domain.IsRepositoryError(err):
		var repoErr *domain.RepositoryError
		errors.As(err, &repoErr)
		if domain.IsNotFoundError(repoErr.Unwrap()) {
			statusCode = http.StatusNotFound
			errorCode = "BET_NOT_FOUND"
			message = "bet not found"
			logger.Info("resource not found", append(logFields, zap.String("error_code", errorCode))...)
		} else {
			statusCode = http.StatusInternalServerError
			errorCode = "REPOSITORY_ERROR"
			message = "database operation failed"
			logger.Error("repository error", append(logFields, zap.String("error_code", errorCode))...)
		}

	default:
		statusCode = http.StatusInternalServerError
		errorCode = "INTERNAL_ERROR"
		message = "An internal error occurred"
		logger.Error("unhandled error", append(logFields, zap.String("error_code", errorCode))...)
	}

	sendErrorResponse(w, statusCode, errorCode, message, logger)
}

func sendErrorResponse(w http.ResponseWriter, status int, code, message string, logger *zap.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := ErrorResponse{
		Error: message,
		Code:  code,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("failed to encode error response", zap.Error(err))
	}
}
