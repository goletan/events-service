package config

import (
	"github.com/goletan/config/pkg"
	"github.com/goletan/events/internal/types"
	observability "github.com/goletan/observability/pkg"
	"go.uber.org/zap"
)

var cfg types.EventsConfig

func LoadEventsConfig(obs *observability.Observability) (*types.EventsConfig, error) {
	if err := config.LoadConfig("Events", &cfg, obs.Logger); err != nil {
		obs.Logger.WithContext(map[string]interface{}{
			"step":    "config loading",
			"error":   zap.Error(err),
			"message": "Failed to load events configuration",
		})
		return nil, err
	}

	return &cfg, nil
}
