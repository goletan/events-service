package service

import (
	"context"
	"errors"
	"fmt"
	eventsLib "github.com/goletan/events-library/pkg"
	"github.com/goletan/events-library/proto"
	"github.com/goletan/events-service/internal/events"
	"github.com/goletan/events-service/internal/metrics"
	"github.com/goletan/events-service/internal/producer"
	"github.com/goletan/events-service/internal/types"
	"github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type EventsService struct {
	cfg          *types.EventsConfig
	obs          *observability.Observability
	producer     *producer.Producer
	eventsClient *eventsLib.EventsClient
	grpcServer   *events.GRPCServer
	grpcSrv      *grpc.Server
}

func NewEventsService(cfg *types.EventsConfig, obs *observability.Observability) (*EventsService, error) {
	// Initialize EventsClient
	eventsClient, err := eventsLib.NewEventsClient(obs, cfg.GRPC.Address)
	if err != nil {
		return nil, err
	}

	// Initialize Pulsar producer
	prod, err := producer.NewProducer(cfg, obs)
	if err != nil {
		obs.Logger.Error("Failed to initialize producer", zap.Error(err))
		return nil, err
	}

	// Create GRPCServer instance
	grpcServer := events.NewGRPCServer(prod.Producer, obs)

	return &EventsService{
		cfg:          cfg,
		obs:          obs,
		producer:     prod,
		eventsClient: eventsClient,
		grpcServer:   grpcServer,
	}, nil
}

func (s *EventsService) Run(ctx context.Context) error {
	s.obs.Logger.Info("Initializing Events Service...")

	metrics.InitMetrics(s.obs)

	// Start gRPC server
	if err := s.startGRPCServer(ctx); err != nil {
		s.obs.Logger.Error("Failed to start gRPC server", zap.Error(err))
		return err
	}

	// Start producer
	go s.producer.Start(ctx)

	return nil
}

func (s *EventsService) startGRPCServer(ctx context.Context) error {
	if s.grpcSrv != nil {
		return fmt.Errorf("gRPC server is already running")
	}

	if s.cfg.GRPC.Address == "" {
		return fmt.Errorf("gRPC address is not configured")
	}

	s.obs.Logger.Info("Initializing gRPC server", zap.String("address", s.cfg.GRPC.Address))

	s.grpcSrv = grpc.NewServer()

	// Register GRPCServer with gRPC runtime
	proto.RegisterEventServiceServer(s.grpcSrv, s.grpcServer)

	listener, err := net.Listen("tcp", s.cfg.GRPC.Address)
	if err != nil {
		s.obs.Logger.Error("Failed to listen for gRPC", zap.Error(err))
		return err
	}

	// Run gRPC server in separate goroutine
	go func() {
		if err := s.grpcSrv.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			s.obs.Logger.Error("gRPC server stopped unexpectedly", zap.Error(err))
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()
	s.obs.Logger.Info("Shutting down gRPC server...")
	s.grpcSrv.GracefulStop() // Perform clean shutdown.

	return nil
}
