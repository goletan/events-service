package service

import (
	servicesTypes "github.com/goletan/services-library/shared/types"
	"log"
)

type EventsService struct {
	ServiceName    string
	ServiceAddress string
	Ports          []servicesTypes.ServicePort
	Tags           map[string]string
}

func (e *EventsService) Name() string {
	return e.ServiceName
}

func (e *EventsService) Address() string {
	return e.ServiceAddress
}

func (e *EventsService) Initialize() error {
	// Logic to initialize the service
	log.Print("Initializing fake core service...")
	return nil
}

func (e *EventsService) Start() error {
	// Logic to start the service
	log.Print("Start fake core service...")
	return nil
}

func (e *EventsService) Stop() error {
	// Logic to stop the service
	log.Print("Stop fake core service...")
	return nil
}
