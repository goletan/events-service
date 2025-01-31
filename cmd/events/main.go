package main

import (
	"context"
	"github.com/goletan/events-service/internal/events"
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

	// Create and run the Events Service
	newEvents, err := events.NewEvents()
	if err != nil {
		log.Fatal("Failed to create events service", err)
		return
	}

	if err = newEvents.Run(shutdownCtx); err != nil {
		newEvents.Observability.Logger.Fatal("Failed to run Events Service", zap.Error(err))
	}

	newEvents.Observability.Logger.Info("Events Service is running...")
	<-shutdownCtx.Done()
	newEvents.Observability.Logger.Info("Shutting down Events Service...")
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
