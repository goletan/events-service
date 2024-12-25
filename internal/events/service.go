package events

import (
	"fmt"
	observability "github.com/goletan/observability-library/pkg"
	services "github.com/goletan/services-library/pkg"
	"github.com/goletan/services-library/shared/types"
)

type EventsService struct {
	endpoint types.ServiceEndpoint
}

func (e *EventsService) Discover(obs *observability.Observability) ([]types.ServiceEndpoint, error) {
	//TODO implement me
}

func NewService(endpoint types.ServiceEndpoint) *EventsService {
	return &EventsService{endpoint: endpoint}
}

// Implement the Service interface...
func (e *EventsService) Name() string { return e.endpoint.Name }

func (e *EventsService) Initialize() error {
	fmt.Printf("Initializing Events Service at %s\n", e.endpoint.Address)
	return nil
}
func (e *EventsService) Start() error {
	fmt.Printf("Starting Events Service at %s\n", e.endpoint.Address)
	return nil
}
func (e *EventsService) Stop() error {
	fmt.Printf("Stopping Events Service at %s\n", e.endpoint.Address)
	return nil
}

func RegisterFactory(services *services.Services) {
	services.RegisterFactory("events-service-service", func(endpoint types.ServiceEndpoint) types.Service {
		return NewService(endpoint)
	})
}
