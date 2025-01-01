package producer

import (
	"context"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/goletan/events-service/internal/types"
	"github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
)

type Producer struct {
	Client   pulsar.Client
	Producer pulsar.Producer
	cfg      *types.EventsConfig
	obs      *observability.Observability
}

// NewProducer initializes a new Pulsar producer.
func NewProducer(cfg *types.EventsConfig, obs *observability.Observability) (*Producer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: cfg.Pulsar.Broker.URL,
	})
	if err != nil {
		obs.Logger.Error("Failed to create Pulsar client", zap.Error(err))
		return nil, err
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: cfg.Event.Producer.Topic,
	})
	if err != nil {
		obs.Logger.Error("Failed to create Pulsar producer", zap.Error(err))
		client.Close() // Clean up client on failure
		return nil, err
	}

	return &Producer{
		Client:   client,
		Producer: producer,
		cfg:      cfg,
		obs:      obs,
	}, nil
}

func (p *Producer) Start(ctx context.Context) {
	p.obs.Logger.Info("Starting producer...")
	producer, err := p.Client.CreateProducer(pulsar.ProducerOptions{
		Topic: p.cfg.Event.Producer.Topic,
	})
	if err != nil {
		p.obs.Logger.Fatal("Failed to create producer", zap.Error(err))
		return
	}
	defer producer.Close()
}

// SendMessage sends a message to the configured topic with metrics and tracing.
func (p *Producer) SendMessage(ctx context.Context, eventType string, payload []byte) error {
	// Start a span for tracing
	ctx, span := p.obs.Tracer.Start(ctx, "send-message")
	defer span.End()

	msg := pulsar.ProducerMessage{
		Payload: payload,
		Properties: map[string]string{
			"event_type": eventType,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		},
	}

	_, err := p.Producer.Send(ctx, &msg)
	if err != nil {
		p.obs.Logger.Error("Failed to send message", zap.Error(err), zap.String("event_type", eventType))
		return err
	}

	// Log and record metrics on success
	p.obs.Logger.Info("Message successfully sent", zap.String("event_type", eventType))
	return nil
}

// Stop gracefully shuts down the producer.
func (p *Producer) Stop() {
	p.obs.Logger.Info("Stopping producer...")
	if p.Producer != nil {
		p.Producer.Close()
	}
	if p.Client != nil {
		p.Client.Close()
	}
	p.obs.Logger.Info("Producer stopped.")
}
