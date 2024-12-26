package events

import (
	"context"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/goletan/events-service/internal/types"
	observability "github.com/goletan/observability-library/pkg"
	"go.uber.org/zap"
)

type Service struct {
	name   string
	obs    *observability.Observability
	client pulsar.Client
	cfg    *types.EventsConfig
}

func NewService(name string, obs *observability.Observability, cfg *types.EventsConfig) *Service {
	return &Service{name: name, obs: obs, cfg: cfg}
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) Initialize() error {
	s.obs.Logger.Info("Initializing Events Service...")
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: s.cfg.Pulsar.Broker.URL,
	})
	if err != nil {
		s.obs.Logger.Error("Failed to create Pulsar client", zap.Error(err))
		return err
	}
	s.client = client
	return nil
}

func (s *Service) Start() error {
	s.obs.Logger.Info("Starting Events Service...")
	go s.runProducer()
	go s.runConsumer()
	return nil
}

func (s *Service) Stop() error {
	s.obs.Logger.Info("Stopping Events Service...")
	if s.client != nil {
		s.client.Close()
		s.obs.Logger.Info("Pulsar client closed")
	}
	return nil
}

func (s *Service) runProducer() {
	producer, err := s.client.CreateProducer(pulsar.ProducerOptions{
		Topic: s.cfg.Event.Producer.Topic,
	})
	if err != nil {
		s.obs.Logger.Error("Failed to create Pulsar producer", zap.Error(err))
		return
	}
	defer producer.Close()

	msg := pulsar.ProducerMessage{Payload: []byte("Hello, Goletan Events!")}
	if _, err := producer.Send(context.Background(), &msg); err != nil {
		s.obs.Logger.Error("Failed to send message", zap.Error(err))
	} else {
		s.obs.Logger.Info("Message successfully sent to Pulsar topic.")
	}
}

func (s *Service) runConsumer() {
	consumer, err := s.client.Subscribe(pulsar.ConsumerOptions{
		Topic:            s.cfg.Event.Producer.Topic,
		SubscriptionName: s.cfg.Event.Consumer.Subscription.Name,
		Type:             pulsar.Shared,
	})
	if err != nil {
		s.obs.Logger.Fatal("Failed to create Pulsar consumer", zap.Error(err))
		return
	}
	defer consumer.Close()

	for {
		ctx, span := s.obs.Tracer.Start(context.Background(), "consume-event")
		msg, err := consumer.Receive(ctx)
		if err != nil {
			s.obs.Logger.Error("Failed to receive message", zap.Error(err))
			continue
		}

		s.obs.Logger.Info("Message received", zap.ByteString("payload", msg.Payload()))
		if err := consumer.Ack(msg); err != nil {
			s.obs.Logger.Error("Failed to acknowledge message", zap.Error(err))
		} else {
			s.obs.Logger.Info("Message acknowledged successfully", zap.String("payload", string(msg.Payload())))
		}

		span.End()
	}
}
