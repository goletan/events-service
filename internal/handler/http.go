package handler

import (
	"encoding/json"
	"github.com/goletan/events-library/shared/types"
	observability "github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
	"net/http"
)

type HTTPHandler struct {
	obs *observability.Observability
}

func NewHTTPHandler(obs *observability.Observability) *HTTPHandler {
	return &HTTPHandler{obs: obs}
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event types.Event
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		h.obs.Logger.Error("Failed to decode event", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	h.obs.Logger.Info("Received event via HTTP", zap.String("event_type", event.Type), zap.String("payload", string(event.Payload)))
	w.WriteHeader(http.StatusOK)
}
