// /events/internal/config/config.go
package config

import (
	config "github.com/goletan/config/pkg"
	"github.com/goletan/events/internal/types"
	"go.uber.org/zap"
)

var cfg types.EventsConfig

func LoadEventsConfig(logger *zap.Logger) (*types.EventsConfig, error) {
	if err := config.LoadConfig("Events", &cfg, logger); err != nil {
		logger.Error(
			"Failed to load events configuration",
			zap.Error(err),
			zap.Any("context", map[string]interface{}{"step": "config loading"}),
		)
		return nil, err
	}

	return &cfg, nil
}
