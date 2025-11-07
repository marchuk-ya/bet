package handler

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

func sendJSON(w http.ResponseWriter, status int, data interface{}, logger *zap.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("failed to encode response", zap.Error(err))
	}
}
