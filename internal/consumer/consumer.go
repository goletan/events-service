package consumer

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/goletan/events-service/internal/metrics"
	"github.com/goletan/events-service/internal/types"
	"github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
)

type Consumer struct {
	client     pulsar.Client
	cfg        *types.EventsConfig
	obs        *observability.Observability
	processor  func(msg pulsar.Message) error // Message processor
	subscriber pulsar.Consumer
}

func NewConsumer(cfg *types.EventsConfig, obs *observability.Observability) (*Consumer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: cfg.Pulsar.Broker.URL,
	})
	if err != nil {
		obs.Logger.Error("Failed to create Pulsar client", zap.Error(err))
		return nil, err
	}

	return &Consumer{
		client: client,
		cfg:    cfg,
		obs:    obs,
		processor: func(msg pulsar.Message) error {
			// Default processor simply logs the message
			obs.Logger.Info("Processing message", zap.ByteString("payload", msg.Payload()))
			return nil
		},
	}, nil
}

func (c *Consumer) Start(ctx context.Context) {
	c.obs.Logger.Info("Starting consumer...")
	consumer, err := c.client.Subscribe(pulsar.ConsumerOptions{
		Topic:            c.cfg.Event.Producer.Topic,
		SubscriptionName: c.cfg.Event.Consumer.Subscription.Name,
		Type:             pulsar.Shared, // Change this based on your use case
	})
	if err != nil {
		c.obs.Logger.Fatal("Failed to create consumer", zap.Error(err))
		return
	}
	defer consumer.Close()

	c.subscriber = consumer

	// Start consuming messages
	for {
		select {
		case <-ctx.Done():
			return
		default:
			msg, err := consumer.Receive(ctx)
			if err != nil {
				c.obs.Logger.Error("Failed to receive message", zap.Error(err))
				metrics.IncrementEventFailed(msg.Topic())
				continue
			}
			c.obs.Logger.Info("Received message", zap.ByteString("payload", msg.Payload()))
			if err := c.processor(msg); err != nil {
				c.obs.Logger.Error("Message processing failed", zap.Error(err))
			} else {
				c.obs.Logger.Info("Acknowledging message", zap.ByteString("payload", msg.Payload()))
				err := consumer.Ack(msg)
				if err != nil {
					c.obs.Logger.Error("Failed to acknowledge message", zap.Error(err))
					metrics.IncrementEventFailed(msg.Topic())
					return
				}
				metrics.IncrementEventProcessed(msg.Topic())
			}
		}
	}
}

// SetProcessor allows setting a custom message processor.
func (c *Consumer) SetProcessor(processor func(msg pulsar.Message) error) {
	c.processor = processor
}
