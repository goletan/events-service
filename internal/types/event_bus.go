package types

import (
	"time"
)

// EventConfig
type EventsConfig struct {
	EventBus struct {
		BulkheadCapacity   int `mapstructure:"bulkhead_capacity"`
		DefaultRetryPolicy struct {
			MaxRetries int           `mapstructure:"max_retries"`
			Backoff    time.Duration `mapstructure:"backoff"`
		} `mapstructure:"default_retry_policy"`
		DLQ struct {
			MaxSize int `mapstructure:"max_size"`
		} `mapstructure:"dlq"`
	} `mapstructure:"event_bus"`
}

// RetryPolicy represents the configuration for retrying event delivery.
type RetryPolicy struct {
	MaxRetries int           // Maximum number of retries
	Backoff    time.Duration // Time to wait before retrying
}

// DeadLetterQueue represents the configuration for handling failed events.
type DeadLetterQueue struct {
	Queue   chan Event
	MaxSize int
}

// Event represents an event with a name, payload, and priority.
type Event struct {
	Name     string
	Payload  interface{}
	Priority int // 1 = High, 2 = Medium, 3 = Low
}

// FilterFunc defines a function signature for filtering events.
type FilterFunc func(event Event) bool
