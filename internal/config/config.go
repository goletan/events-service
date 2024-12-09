package config

import (
	"github.com/goletan/config/pkg"
	"github.com/goletan/events/internal/types"
	"github.com/goletan/observability/shared/logger"
	"go.uber.org/zap"
)

var cfg types.EventsConfig

func LoadEventsConfig(log *logger.ZapLogger) (*types.EventsConfig, error) {
	if err := config.LoadConfig("Events", &cfg, log); err != nil {
		log.WithContext(map[string]interface{}{
			"step":    "config loading",
			"error":   zap.Error(err),
			"message": "Failed to load events configuration",
		})
		return nil, err
	}

	return &cfg, nil
}
