// Xomoi-Core: Sovereign Edge Node
// Copyright (C) 2026 Adrish Bora (@code-grey) & Simanjit Hujuri (@code-zephyrus)
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

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
