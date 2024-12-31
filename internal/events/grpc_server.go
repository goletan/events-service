package events

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/goletan/events-library/proto"
	"github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
)

type GRPCServer struct {
	proto.UnimplementedEventServiceServer
	producer pulsar.Producer
	obs      *observability.Observability
}

func NewGRPCServer(producer pulsar.Producer, obs *observability.Observability) *GRPCServer {
	return &GRPCServer{
		producer: producer,
		obs:      obs,
	}
}

func (s *GRPCServer) SendEvent(ctx context.Context, req *proto.EventRequest) (*proto.EventResponse, error) {
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
	return &proto.EventResponse{Status: "success"}, nil
}
