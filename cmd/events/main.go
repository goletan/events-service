// /events/cmd/events/main.go
package main

import (
	"github.com/goletan/events/internal/events"
	"github.com/goletan/observability/logger"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger()
	log := logger.GetInstance()
	cfg, err := events.LoadEventsConfig(log)
	if err != nil {
		logger.Fatal("Failed to load kernel configuration", zap.Error(err))
	}

	eventBus = events.NewEventBus(cfg)
	logger.Info("Event bus initialized")
}
