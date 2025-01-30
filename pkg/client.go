package pkg

import (
	"context"
	"github.com/goletan/events-service/internal/client/strategies"
	"github.com/goletan/events-service/internal/config"
	"github.com/goletan/events-service/internal/metrics"
	"github.com/goletan/events-service/internal/types"
	sharedTypes "github.com/goletan/events-service/shared/types"
	observability "github.com/goletan/observability-library/pkg"
	resilience "github.com/goletan/resilience-library/pkg"
	resTypes "github.com/goletan/resilience-library/shared/types"
	"github.com/sony/gobreaker/v2"
	"go.uber.org/zap"
	"time"
)

type EventsClient struct {
	cfg        *types.EventsConfig
	obs        *observability.Observability
	strategy   strategies.Strategy
	resilience *resilience.DefaultResilienceService
	metrics    *metrics.EventsMetrics
}

func NewEventsClient(obs *observability.Observability) (*EventsClient, error) {
	// Load configuration
	cfg, err := config.LoadEventsConfig(obs)
	if err != nil {
		obs.Logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize resilience service
	res := resilience.NewResilienceService(
		cfg.ServiceName,
		obs,
		func(err error) bool { return true }, // Retry on all errors for now
		&resTypes.CircuitBreakerCallbacks{
			OnOpen: func(name string, from, to gobreaker.State) {
				obs.Logger.Warn("Circuit breaker opened", zap.String("name", name))
			},
			OnClose: func(name string, from, to gobreaker.State) {
				obs.Logger.Info("Circuit breaker closed", zap.String("name", name))
			},
		},
	)

	// Initialize metrics
	met := metrics.InitMetrics(obs)
	met.Register()

	strategy, _ := func() (strategies.Strategy, error) {
		switch cfg.Client.Strategy {
		case "grpc":
			obs.Logger.Info("Using gRPC strategy")
			return strategies.NewGRPCStrategy(cfg.GRPC.Address, obs)
		case "http":
			obs.Logger.Info("Using HTTP strategy")
			return strategies.NewHTTPStrategy(cfg.HTTP.Address, obs)
		default:
			obs.Logger.Warn("Invalid strategy specified, defaulting to gRPC.", zap.String("strategy", cfg.Client.Strategy))
			return strategies.NewGRPCStrategy(cfg.GRPC.Address, obs)
		}
	}()

	return &EventsClient{
		cfg:        cfg,
		obs:        obs,
		strategy:   strategy,
		resilience: res,
		metrics:    met,
	}, nil
}

func (ec *EventsClient) SendEvent(ctx context.Context, event *sharedTypes.Event) error {
	start := time.Now()

	// Use resilience and metrics
	err := ec.resilience.ExecuteWithRetry(ctx, func() error {
		return ec.strategy.Send(ctx, event)
	})

	duration := time.Since(start).Seconds()
	if err != nil {
		ec.metrics.IncrementFailed(event.Type)
		ec.obs.Logger.Error("Failed to send event", zap.Error(err), zap.String("event_type", event.Type))
		return err
	}

	ec.metrics.IncrementPublished(event.Type)
	ec.metrics.ObserveLatency(event.Type, duration)

	ec.obs.Logger.Info("Event sent successfully", zap.String("event_type", event.Type), zap.Float64("latency_seconds", duration))
	return nil
}
