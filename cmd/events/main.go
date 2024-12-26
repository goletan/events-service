package main

import (
	"context"
	"github.com/goletan/events-service/internal/config"
	"github.com/goletan/events-service/internal/events"
	observabilityLib "github.com/goletan/observability-library/pkg"
	servicesLib "github.com/goletan/services-library/pkg"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Graceful shutdown setup
	shutdownCtx, shutdownCancel := setupShutdownContext()
	defer shutdownCancel()

	// Initialize observabilityLib
	obs := initializeObservability()

	// Load configuration
	cfg, err := config.LoadEventsConfig(obs)
	if err != nil {
		obs.Logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize the Services Library
	services, err := servicesLib.NewServices(obs)
	if err != nil {
		obs.Logger.Fatal("Failed to initialize services library", zap.Error(err))
	}

	// Create and register the Events Service
	eventsService := events.NewService(cfg.ServiceName, obs, cfg)
	if err := services.Register(eventsService); err != nil {
		obs.Logger.Fatal("Failed to register Events Service", zap.Error(err))
	}

	// Initialize and start the service
	if err := services.InitializeAll(shutdownCtx); err != nil {
		obs.Logger.Fatal("Failed to initialize services", zap.Error(err))
	}

	if err := services.StartAll(shutdownCtx); err != nil {
		obs.Logger.Fatal("Failed to start services", zap.Error(err))
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

// initializeObservability initializes the observabilityLib-library components.
func initializeObservability() *observabilityLib.Observability {
	obs, err := observabilityLib.NewObserver()
	if err != nil {
		log.Fatal("Failed to initialize observabilityLib-library", err)
	}
	return obs
}
