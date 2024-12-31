package client

import (
	"context"
	proto2 "github.com/goletan/events-library/proto"
	"time"

	"github.com/goletan/observability-library/pkg"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// EventClientInterface defines an interface for the EventsClient to allow easier testing/mocking
type EventClientInterface interface {
	SendEvent(ctx context.Context, eventType, payload string) (string, error)
	Close() error
}

// EventsClient is the concrete implementation of EventClientInterface for interacting with the Event Service.
type EventsClient struct {
	client proto2.EventServiceClient
	conn   *grpc.ClientConn // store the connection for cleanup
	obs    *observability.Observability
}

// NewEventsClient creates a new instance of EventsClient and sets up the gRPC connection.
// For production, consider using secure credentials instead of insecure.
func NewEventsClient(address string, obs *observability.Observability) (*EventsClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		if obs != nil {
			obs.Logger.Error("Failed to connect to Events Service", zap.Error(err))
		}
		return nil, errors.Wrap(err, "failed to connect to Events Service")
	}

	return &EventsClient{
		client: proto2.NewEventServiceClient(conn),
		conn:   conn,
		obs:    obs,
	}, nil
}

// SendEvent sends an event to the Event Service with a context-based timeout and proper logging.
func (ec *EventsClient) SendEvent(ctx context.Context, eventType string, payload string) (string, error) {
	// Set a timeout to avoid hanging
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &proto2.EventRequest{
		EventType: eventType,
		Payload:   payload,
	}

	if ec.obs != nil {
		ec.obs.Logger.Info("Sending event", zap.String("event_type", eventType), zap.String("payload", payload))
	}

	res, err := ec.client.SendEvent(ctx, req)
	if err != nil {
		if ec.obs != nil {
			ec.obs.Logger.Error("Failed to send event", zap.Error(err), zap.String("event_type", eventType))
		}
		return "", errors.Wrap(err, "failed to send event")
	}

	if ec.obs != nil {
		ec.obs.Logger.Info("Event sent successfully", zap.String("status", res.Status))
	}
	return res.Status, nil
}

// Close cleans up the gRPC connection to prevent resource leaks.
func (ec *EventsClient) Close() error {
	if err := ec.conn.Close(); err != nil {
		if ec.obs != nil {
			ec.obs.Logger.Error("Failed to close gRPC connection", zap.Error(err))
		}
		return errors.Wrap(err, "failed to close gRPC connection")
	}
	return nil
}
