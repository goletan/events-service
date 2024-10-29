package types

import (
	"time"
)

// Event represents an event with a name, payload, and priority.
type Event struct {
	Name     string
	Payload  interface{}
	Priority int // 1 = High, 2 = Medium, 3 = Low
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

// FilterFunc defines a function signature for filtering events.
type FilterFunc func(event Event) bool
