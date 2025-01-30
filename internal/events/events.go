package events

import (
	"context"
	"errors"
	"fmt"
	"github.com/goletan/events-service/internal/metrics"
	"github.com/goletan/events-service/internal/producer"
	"github.com/goletan/events-service/internal/types"
	"github.com/goletan/events-service/proto"
	"github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type EventsService struct {
	cfg        *types.EventsConfig
	obs        *observability.Observability
	producer   *producer.Producer
	grpcServer *GRPCServer
	grpcSrv    *grpc.Server
}

func NewEventsService(obs *observability.Observability, cfg *types.EventsConfig) (*EventsService, error) {
	// Initialize Pulsar producer
	prod, err := producer.NewProducer(cfg, obs)
	if err != nil {
		obs.Logger.Error("Failed to initialize producer", zap.Error(err))
		return nil, err
	}

	// Create GRPCServer instance
	grpcServer := NewGRPCServer(prod.Producer, obs)

	return &EventsService{
		cfg:        cfg,
		obs:        obs,
		producer:   prod,
		grpcServer: grpcServer,
	}, nil
}

func (es *EventsService) Run(ctx context.Context) error {
	es.obs.Logger.Info("Initializing Events Service...")

	metrics.InitMetrics(es.obs)

	// Start gRPC server
	if err := es.startServer(ctx); err != nil {
		es.obs.Logger.Error("Failed to start gRPC server", zap.Error(err))
		return err
	}

	// Start producer
	go es.producer.Start(ctx)

	return nil
}

func (es *EventsService) startServer(ctx context.Context) error {
	if es.grpcSrv != nil {
		return fmt.Errorf("gRPC server is already running")
	}

	if es.cfg.GRPC.Address == "" {
		return fmt.Errorf("gRPC address is not configured")
	}

	es.obs.Logger.Info("Initializing gRPC server", zap.String("address", es.cfg.GRPC.Address))

	es.grpcSrv = grpc.NewServer()

	// Register GRPCServer with gRPC runtime
	proto.RegisterEventServiceServer(es.grpcSrv, es.grpcServer)

	listener, err := net.Listen("tcp", es.cfg.GRPC.Address)
	if err != nil {
		es.obs.Logger.Error("Failed to listen for gRPC", zap.Error(err))
		return err
	}

	// Run gRPC server in separate goroutine
	go func() {
		if err := es.grpcSrv.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			es.obs.Logger.Error("gRPC server stopped unexpectedly", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	es.obs.Logger.Info("Shutting down gRPC server...")
	es.grpcSrv.GracefulStop() // Perform clean shutdown.

	return nil
}
