package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/goletan/events-service/internal/consumer"
	"github.com/goletan/events-service/internal/metrics"
	"github.com/goletan/events-service/internal/producer"
	"github.com/goletan/events-service/internal/types"
	"github.com/goletan/events-service/proto" // Import your generated gRPC files
	"github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
)

type EventsService struct {
	proto.UnimplementedEventServiceServer
	cfg      *types.EventsConfig
	obs      *observability.Observability
	producer *producer.Producer
	consumer *consumer.Consumer
	grpcSrv  *grpc.Server
}

func NewEventsService(cfg *types.EventsConfig, obs *observability.Observability) *EventsService {
	return &EventsService{
		cfg: cfg,
		obs: obs,
	}
}

func (s *EventsService) Run(ctx context.Context) error {
	s.obs.Logger.Info("Initializing Events Service...")

	metrics.InitMetrics(s.obs)

	// Initialize producer
	prod, err := producer.NewProducer(s.cfg, s.obs)
	if err != nil {
		s.obs.Logger.Error("Failed to initialize producer", zap.Error(err))
		return err
	}
	s.producer = prod

	// Initialize consumer
	cons, err := consumer.NewConsumer(s.cfg, s.obs)
	if err != nil {
		s.obs.Logger.Error("Failed to initialize consumer", zap.Error(err))
		return err
	}
	s.consumer = cons

	// Start gRPC server
	if err := s.startGRPCServer(ctx); err != nil {
		s.obs.Logger.Error("Failed to start gRPC server", zap.Error(err))
		return err
	}

	// Start producer and consumer
	go s.producer.Start(ctx)
	go s.consumer.Start(ctx)

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

	// Register EventService with gRPC server
	proto.RegisterEventServiceServer(s.grpcSrv, s)
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

// SendEvent Implement the gRPC interface
func (s *EventsService) SendEvent(ctx context.Context, req *proto.EventRequest) (*proto.EventResponse, error) {
	s.obs.Logger.Info("Received gRPC event", zap.String("event_type", req.EventType))
	// TODO: Add event processing logic here
	return &proto.EventResponse{Status: "success"}, nil
}
