package config

import (
	"github.com/goletan/config/pkg"
	"github.com/goletan/events/internal/types"
	observability "github.com/goletan/observability/pkg"
	"go.uber.org/zap"
)

var cfg types.EventsConfig

func LoadEventsConfig(obs *observability.Observability) (*types.EventsConfig, error) {
	if err := config.LoadConfig("Events", &cfg, obs); err != nil {
		obs.Logger.Error(
			"Failed to load events configuration",
			zap.Error(err),
			zap.Any("context", map[string]interface{}{"step": "config loading"}),
		)
		return nil, err
	}

	return &cfg, nil
}
