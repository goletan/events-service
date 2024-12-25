package main

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/goletan/events/internal/config"
	"github.com/goletan/events/internal/types"
	observability "github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Setup context and signal handling for graceful shutdown
	shutdownCtx, shutdownCancel := setupShutdownContext()
	defer shutdownCancel()

	// Initialize observability-library
	obs := initializeObservability()

	// Initialize configuration
	cfg, err := config.LoadEventsConfig(obs)
	if err != nil {
		return // Exit if client initialization fails
	}

	// Initialize Pulsar client
	client, err := initializePulsarClient(obs, cfg)
	if err != nil {
		return // Exit if client initialization fails
	}
	defer client.Close()

	obs.Logger.Info("Pulsar client initialized. Starting Events Service...")

	// Start producer and consumer routines
	startEventProcessing(obs, cfg, client)

	// Wait for shutdown signal
	obs.Logger.Info("Events Service is running...")
	<-shutdownCtx.Done()
	obs.Logger.Info("Events Service shutting down...")
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

// initializeObservability initializes the observability-library components.
func initializeObservability() *observability.Observability {
	obs, err := observability.NewObserver()
	if err != nil {
		log.Fatal("Failed to initialize observability-library", err)
	}
	return obs
}

// initializePulsarClient creates and returns a Pulsar client.
func initializePulsarClient(obs *observability.Observability, cfg *types.EventsConfig) (pulsar.Client, error) {
	obs.Logger.Info("Creating Pulsar Client...")
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: cfg.Events.Pulsar.Broker.URL,
	})
	if err != nil {
		obs.Logger.Error("Failed to create Pulsar client.", zap.Error(err))
		return nil, err
	}
	return client, nil
}

// startEventProcessing starts producer and consumer routines.
func startEventProcessing(obs *observability.Observability, cfg *types.EventsConfig, client pulsar.Client) {
	go runProducer(obs, cfg, client)
	go runConsumer(obs, cfg, client)
}

func runProducer(obs *observability.Observability, cfg *types.EventsConfig, client pulsar.Client) {
	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: cfg.Events.Event.Producer.Topic,
	})
	if err != nil {
		obs.Logger.Error("Failed to create Pulsar producer.", zap.Error(err))
		return
	}
	defer producer.Close()

	msg := pulsar.ProducerMessage{
		Payload: []byte("Hello, Goletan Events!"),
	}

	if _, err := producer.Send(context.Background(), &msg); err != nil {
		obs.Logger.Error("Failed to send message.", zap.Error(err))
	} else {
		obs.Logger.Info("Message successfully sent to Pulsar topic.")
	}
}

func runConsumer(obs *observability.Observability, cfg *types.EventsConfig, client pulsar.Client) {
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            cfg.Events.Event.Producer.Topic,
		SubscriptionName: cfg.Events.Event.Consumer.Subscription.Name,
		Type:             pulsar.Shared,
	})
	if err != nil {
		obs.Logger.Fatal("Failed to create Pulsar consumer.", zap.Error(err))
		return
	}
	defer consumer.Close()

	for {
		ctx, span := obs.Tracer.Start(context.Background(), "consume-event")
		msg, err := consumer.Receive(ctx)
		if err != nil {
			obs.Logger.Error("Failed to receive message.", zap.Error(err))
			continue
		}

		log.Printf("Message received.")
		if err := consumer.Ack(msg); err != nil {
			obs.Logger.Error("Failed to acknowledge message.", zap.Error(err))
		}

		span.End() // End trace
	}
}
