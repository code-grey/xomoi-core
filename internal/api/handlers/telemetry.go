package handlers

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/code-grey/xomoi-core/internal/api/response"
	"github.com/code-grey/xomoi-core/internal/repository"
)

type TelemetryHandler struct {
	tsdb repository.TelemetryRepository
}

func NewTelemetryHandler(tsdb repository.TelemetryRepository) *TelemetryHandler {
	return &TelemetryHandler{tsdb: tsdb}
}

func (h *TelemetryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	mac := r.PathValue("mac")
	timeframe := r.URL.Query().Get("timeframe")

	// Default to 1 hour
	duration := -1 * time.Hour
	switch timeframe {
	case "3h":
		duration = -3 * time.Hour
	case "6h":
		duration = -6 * time.Hour
	case "12h":
		duration = -12 * time.Hour
	case "7d":
		duration = -7 * 24 * time.Hour
	case "30d":
		duration = -30 * 24 * time.Hour
	}

	since := time.Now().Add(duration)

	history, err := h.tsdb.GetDeviceHistory(r.Context(), mac, since)
	if err != nil {
		slog.Error("Failed to fetch telemetry history", "error", err, "mac", mac)
		response.Error(w, http.StatusInternalServerError, "Failed to fetch history")
		return
	}

	if len(history) == 0 {
		response.JSON(w, http.StatusOK, []interface{}{})
		return
	}

	response.JSON(w, http.StatusOK, history)
}
