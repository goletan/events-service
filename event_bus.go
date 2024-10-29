// /events/event_bus.go
package events

import (
	"log"
	"sync"
	"time"

	"github.com/goletan/events/types"
)

// EventBus manages the publishing and subscription of events in a thread-safe manner.
type EventBus struct {
	subscribers        map[string]map[int][]*Subscriber
	mu                 sync.RWMutex
	shutdown           chan struct{}
	defaultRetryPolicy types.RetryPolicy
	bulkheads          map[string]chan types.Event // Bulkhead channels to isolate event types
	bulkheadCapacity   int                         // Capacity for each bulkhead
	dlq                *types.DeadLetterQueue      // Dead-letter queue for failed events
}

// Subscriber represents a subscriber with a channel and an optional filter.
type Subscriber struct {
	ch     chan types.Event
	filter types.FilterFunc
}

// NewEventBus creates a new instance of EventBus with a default retry policy and dead-letter queue.
func NewEventBus(bulkheadCapacity, dlqMaxSize int) *EventBus {
	return &EventBus{
		subscribers: make(map[string]map[int][]*Subscriber),
		shutdown:    make(chan struct{}),
		defaultRetryPolicy: types.RetryPolicy{
			MaxRetries: 3,
			Backoff:    100 * time.Millisecond,
		},
		bulkheads:        make(map[string]chan types.Event),
		bulkheadCapacity: bulkheadCapacity,
		dlq: &types.DeadLetterQueue{
			Queue:   make(chan types.Event, dlqMaxSize),
			MaxSize: dlqMaxSize,
		},
	}
}

// Publish sends an event to all subscribers of that event type, with automatic retry.
func (eb *EventBus) Publish(event types.Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	log.Printf("Publishing event: %s with priority: %d", event.Name, event.Priority)

	// Use bulkhead to prevent overloading
	bulkhead, exists := eb.bulkheads[event.Name]
	if !exists {
		bulkhead = make(chan types.Event, eb.bulkheadCapacity)
		eb.bulkheads[event.Name] = bulkhead
		go eb.processBulkhead(event.Name, bulkhead)
	}

	select {
	case bulkhead <- event:
		log.Printf("Event added to bulkhead: %s", event.Name)
	default:
		log.Printf("Bulkhead full, dropping event: %s", event.Name)
	}
}

// processBulkhead processes events from a bulkhead channel.
func (eb *EventBus) processBulkhead(eventName string, bulkhead chan types.Event) {
	for {
		select {
		case <-eb.shutdown:
			log.Printf("Shutting down bulkhead processor for event: %s", eventName)
			return
		case event := <-bulkhead:
			if prioritySubs, found := eb.subscribers[event.Name]; found {
				// Publish to subscribers by priority
				for priority := 1; priority <= 3; priority++ {
					if subs, exists := prioritySubs[priority]; exists {
						for _, sub := range subs {
							if !eb.sendWithRetry(sub, event) {
								// Send to DLQ if sending fails after retries
								eb.sendToDLQ(event)
							}
						}
					}
				}
			}
		}
	}
}

// sendWithRetry sends an event to a subscriber, retrying based on the EventBus's retry policy.
// Returns true if the event was sent successfully, false otherwise.
func (eb *EventBus) sendWithRetry(sub *Subscriber, event types.Event) bool {
	for i := 0; i < eb.defaultRetryPolicy.MaxRetries+1; i++ { // +1 to include the initial attempt
		select {
		case sub.ch <- event:
			log.Printf("Event sent successfully: %s after %d attempts", event.Name, i+1)
			return true
		default:
			log.Printf("Subscriber channel full, retrying event: %s (%d/%d)", event.Name, i+1, eb.defaultRetryPolicy.MaxRetries)
			time.Sleep(eb.defaultRetryPolicy.Backoff)
		}
	}
	log.Printf("Failed to send event after %d attempts: %s", eb.defaultRetryPolicy.MaxRetries, event.Name)
	return false
}

// sendToDLQ sends an event to the Dead-Letter Queue for further investigation.
func (eb *EventBus) sendToDLQ(event types.Event) {
	select {
	case eb.dlq.Queue <- event:
		log.Printf("Event sent to Dead-Letter Queue: %s", event.Name)
	default:
		log.Printf("Dead-Letter Queue full, dropping event: %s", event.Name)
	}
}

// Subscribe registers a subscriber for a specific event name with an optional filter and priority level.
func (eb *EventBus) Subscribe(eventName string, priority int, filter types.FilterFunc) chan types.Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	ch := make(chan types.Event, 10)
	subscriber := &Subscriber{
		ch:     ch,
		filter: filter,
	}

	if _, found := eb.subscribers[eventName]; !found {
		eb.subscribers[eventName] = make(map[int][]*Subscriber)
	}
	eb.subscribers[eventName][priority] = append(eb.subscribers[eventName][priority], subscriber)

	log.Printf("Subscriber added for event: %s with priority: %d", eventName, priority)
	return ch
}

// Unsubscribe removes a subscriber from the event bus, cleaning up resources.
func (eb *EventBus) Unsubscribe(eventName string, subscriber chan types.Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if prioritySubs, found := eb.subscribers[eventName]; found {
		for priority, subs := range prioritySubs {
			for i, sub := range subs {
				if sub.ch == subscriber {
					close(sub.ch)
					eb.subscribers[eventName][priority] = append(subs[:i], subs[i+1:]...)
					log.Printf("Subscriber removed for event: %s with priority: %d", eventName, priority)
					break
				}
			}
		}
	}
}

// Shutdown gracefully shuts down the EventBus, closing all subscriber channels.
func (eb *EventBus) Shutdown() {
	close(eb.shutdown)
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for _, prioritySubs := range eb.subscribers {
		for _, subs := range prioritySubs {
			for _, sub := range subs {
				close(sub.ch)
			}
		}
	}

	eb.subscribers = make(map[string]map[int][]*Subscriber)
	log.Println("EventBus has been shut down gracefully")
}
