package strategies

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goletan/events-service/shared/types"
	"github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"time"
)

type HTTPStrategy struct {
	address string
	obs     *observability.Observability
	client  *http.Client
}

// NewHTTPStrategy initializes the HTTP strategy with a given address and observability instance.
func NewHTTPStrategy(address string, obs *observability.Observability) (*HTTPStrategy, error) {
	if address == "" {
		return nil, errors.New("HTTP strategy requires a valid address")
	}

	return &HTTPStrategy{
		address: address,
		obs:     obs,
		client: &http.Client{
			Timeout: 10 * time.Second, // Default timeout
		},
	}, nil
}

// Send sends an event using HTTP POST.
func (h *HTTPStrategy) Send(ctx context.Context, event *types.Event) error {
	url := fmt.Sprintf("%s/events", h.address) // Assume an endpoint "/events" for posting events.

	// Serialize event to JSON
	payload, err := json.Marshal(event)
	if err != nil {
		h.obs.Logger.Error("Failed to serialize event", zap.Error(err), zap.String("event_type", event.Type))
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// Prepare HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payload))
	if err != nil {
		h.obs.Logger.Error("Failed to create HTTP request", zap.Error(err))
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send HTTP request
	h.obs.Logger.Info("Sending event via HTTP", zap.String("url", url), zap.String("event_type", event.Type))
	resp, err := h.client.Do(req)
	if err != nil {
		h.obs.Logger.Error("Failed to send HTTP request", zap.Error(err))
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		h.obs.Logger.Error("Received non-OK response from HTTP server",
			zap.Int("status_code", resp.StatusCode), zap.String("response_body", string(body)))
		return fmt.Errorf("received non-OK response: %d", resp.StatusCode)
	}

	h.obs.Logger.Info("Event sent successfully via HTTP", zap.String("event_type", event.Type))
	return nil
}

// Close is a no-op for HTTP strategy since no persistent connection is maintained.
func (h *HTTPStrategy) Close() error {
	return nil
}
