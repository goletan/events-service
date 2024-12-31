package events

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	pb "github.com/goletan/events-service/proto"
	"github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
)

type GRPCServer struct {
	pb.UnimplementedEventServiceServer
	producer pulsar.Producer
	obs      *observability.Observability
}

func NewGRPCServer(producer pulsar.Producer, obs *observability.Observability) *GRPCServer {
	return &GRPCServer{
		producer: producer,
		obs:      obs,
	}
}

func (s *GRPCServer) SendEvent(ctx context.Context, req *pb.EventRequest) (*pb.EventResponse, error) {
	msg := pulsar.ProducerMessage{
		Key:     req.EventType,
		Payload: []byte(req.Payload),
	}

	_, err := s.producer.Send(ctx, &msg)
	if err != nil {
		s.obs.Logger.Error("Failed to send event", zap.Error(err))
		return nil, err
	}

	s.obs.Logger.Info("Event sent successfully", zap.String("event_type", req.EventType))
	return &pb.EventResponse{Status: "success"}, nil
}
