package service

import (
	"context"
	"github.com/goletan/events-service/internal/events"
	servicesTypes "github.com/goletan/services-library/shared/types"
	"go.uber.org/zap"
)

type EventsService struct {
	ServiceName    string
	ServiceAddress string
	ServiceType    string
	Ports          []servicesTypes.ServicePort
	Tags           map[string]string
	events         *events.Events
}

func NewEventsService(endpoint servicesTypes.ServiceEndpoint) (servicesTypes.Service, error) {
	newEvents, err := events.NewEvents()
	if err != nil || newEvents == nil {
		return nil, err
	}

	return &EventsService{
		ServiceName:    endpoint.Name,
		ServiceAddress: endpoint.Address,
		ServiceType:    "events",
		Ports:          endpoint.Ports,
		Tags:           endpoint.Tags,
		events:         newEvents,
	}, nil
}

func (e *EventsService) Name() string {
	return e.ServiceName
}

func (e *EventsService) Address() string {
	return e.ServiceAddress
}

func (e *EventsService) Type() string {
	return e.ServiceType
}

func (e *EventsService) Metadata() map[string]string {
	return e.Tags
}

func (e *EventsService) Initialize() error {
	if e.events == nil {
		newEvents, err := events.NewEvents()
		if err != nil || newEvents == nil {
			return err
		}
		e.events = newEvents
	}
	e.events.Observability.Logger.Info("Events service initialized")

	return nil
}

func (e *EventsService) Start(ctx context.Context) error {
	e.events.Observability.Logger.Info("Starting core service...")
	err := e.events.Run(ctx)
	if err != nil {
		e.events.Observability.Logger.Fatal("Failed to start core service", zap.Error(err))
		return err
	}
	e.events.Observability.Logger.Info("Events service is running...")

	return nil
}

func (e *EventsService) Stop(ctx context.Context) error {
	e.events.Observability.Logger.Info("Events service shutting down...")
	err := e.events.Stop(ctx)
	if err != nil {
		e.events.Observability.Logger.Error("Failed to shutdown core-service", zap.Error(err))
		return err
	}
	e.events.Observability.Logger.Info("Events service shut down successfully")

	return nil
}
