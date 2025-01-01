package main

import (
	"context"
	"github.com/goletan/events-service/internal/config"
	"github.com/goletan/events-service/internal/service"
	"github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Graceful shutdown context
	shutdownCtx, shutdownCancel := setupShutdownContext()
	defer shutdownCancel()

	// Initialize observability
	obs := initializeObservability()

	// Load configuration
	cfg, err := config.LoadEventsConfig(obs)
	if err != nil {
		obs.Logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Create and run the Events Service
	eventsService, err := service.NewEventsService(cfg, obs)
	if err != nil {
		obs.Logger.Fatal("Failed to create Events Service", zap.Error(err))
		return
	}

	if err := eventsService.Run(shutdownCtx); err != nil {
		obs.Logger.Fatal("Failed to run Events Service", zap.Error(err))
	}

	obs.Logger.Info("Events Service is running...")
	<-shutdownCtx.Done()
	obs.Logger.Info("Shutting down Events Service...")
}

// setupShutdownContext creates a context with signal handling for graceful shutdown.
func setupShutdownContext() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		cancel()
	}()
	return ctx, cancel
}

// initializeObservability initializes observability components.
func initializeObservability() *observability.Observability {
	obs, err := observability.NewObserver()
	if err != nil {
		log.Fatal("Failed to initialize observability", err)
	}
	return obs
}
