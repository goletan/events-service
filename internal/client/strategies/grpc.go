package strategies

import (
	"context"
	"github.com/goletan/events-service/internal/client/serializer"
	"github.com/goletan/events-service/proto"
	"github.com/goletan/events-service/shared/types"
	"github.com/goletan/observability-library/pkg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
)

type GRPCStrategy struct {
	address    string
	obs        *observability.Observability
	client     proto.EventServiceClient
	conn       *grpc.ClientConn
	serializer *serializer.Serializer
}

func NewGRPCStrategy(address string, obs *observability.Observability) (*GRPCStrategy, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		obs.Logger.Error("Failed to connect to Events Service", zap.Error(err))
		return nil, errors.Wrap(err, "failed to connect to Events Service")
	}

	client := proto.NewEventServiceClient(conn)
	return &GRPCStrategy{
		address:    address,
		obs:        obs,
		client:     client,
		conn:       conn,
		serializer: serializer.NewSerializer(),
	}, nil
}

func (g *GRPCStrategy) Send(ctx context.Context, event *types.Event) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Serialize event payload
	data, err := g.serializer.SerializeEvent(event)
	if err != nil {
		g.obs.Logger.Error("Failed to serialize event payload", zap.Error(err))
		return err
	}

	req := &proto.EventRequest{
		EventType: event.Type,
		Payload:   string(data),
	}

	g.obs.Logger.Info("Sending event", zap.String("event_type", event.Type), zap.ByteString("payload", data))

	res, err := g.client.SendEvent(ctx, req)
	if err != nil {
		g.obs.Logger.Error("Failed to send event", zap.Error(err), zap.String("event_type", event.Type))
		return errors.Wrap(err, "failed to send event")
	}

	g.obs.Logger.Info("Event sent successfully", zap.String("status", res.Status))
	return nil
}

// Close cleans up the gRPC connection to prevent resource leaks.
func (g *GRPCStrategy) Close() error {
	if err := g.conn.Close(); err != nil {
		g.obs.Logger.Error("Failed to close gRPC connection", zap.Error(err))
		return errors.Wrap(err, "failed to close gRPC connection")
	}
	return nil
}
