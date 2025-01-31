package events

import (
	"context"
	"errors"
	"fmt"
	"github.com/goletan/events-service/internal/config"
	"github.com/goletan/events-service/internal/metrics"
	"github.com/goletan/events-service/internal/producer"
	"github.com/goletan/events-service/internal/types"
	"github.com/goletan/events-service/proto"
	"github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net"
)

type Events struct {
	Config        *types.EventsConfig
	Observability *observability.Observability
	producer      *producer.Producer
	grpcServer    *GRPCServer
	grpcSrv       *grpc.Server
}

func NewEvents() (*Events, error) {
	obs, err := observability.NewObserver()
	if err != nil {
		log.Fatal("Failed to initialize observability", err)
	}

	// Load configuration
	cfg, err := config.LoadEventsConfig(obs)
	if err != nil {
		obs.Logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize Pulsar producer
	prod, err := producer.NewProducer(cfg, obs)
	if err != nil {
		obs.Logger.Error("Failed to initialize producer", zap.Error(err))
		return nil, err
	}

	// Create GRPCServer instance
	grpcServer := NewGRPCServer(prod.Producer, obs)

	return &Events{
		Config:        cfg,
		Observability: obs,
		producer:      prod,
		grpcServer:    grpcServer,
	}, nil
}

func (es *Events) Run(ctx context.Context) error {
	es.Observability.Logger.Info("Initializing Events Service...")

	metrics.InitMetrics(es.Observability)

	// Start gRPC server
	if err := es.startServer(ctx); err != nil {
		es.Observability.Logger.Error("Failed to start gRPC server", zap.Error(err))
		return err
	}

	// Start producer
	go es.producer.Start(ctx)

	return nil
}

func (es *Events) Stop(ctx context.Context) error {
	es.Observability.Logger.Info("Stopping Events Service...")
	es.producer.Stop()
	es.grpcSrv.Stop()

	return nil
}

func (es *Events) startServer(ctx context.Context) error {
	if es.grpcSrv != nil {
		return fmt.Errorf("gRPC server is already running")
	}

	if es.Config.GRPC.Address == "" {
		return fmt.Errorf("gRPC address is not configured")
	}

	es.Observability.Logger.Info("Initializing gRPC server", zap.String("address", es.Config.GRPC.Address))

	es.grpcSrv = grpc.NewServer()

	// Register GRPCServer with gRPC runtime
	proto.RegisterEventServiceServer(es.grpcSrv, es.grpcServer)

	listener, err := net.Listen("tcp", es.Config.GRPC.Address)
	if err != nil {
		es.Observability.Logger.Error("Failed to listen for gRPC", zap.Error(err))
		return err
	}

	// Run gRPC server in separate goroutine
	go func() {
		if err := es.grpcSrv.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			es.Observability.Logger.Error("gRPC server stopped unexpectedly", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	es.Observability.Logger.Info("Shutting down gRPC server...")
	es.grpcSrv.GracefulStop() // Perform clean shutdown.

	return nil
}
