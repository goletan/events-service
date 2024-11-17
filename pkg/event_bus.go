// /events/pkg/event_bus.go
package events

import (
	"sync"
	"time"

	"github.com/goletan/events/internal/config"
	"github.com/goletan/events/internal/metrics"
	"github.com/goletan/events/internal/types"
	observability "github.com/goletan/observability/pkg"
	"go.uber.org/zap"
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
	observability      *observability.Observability
}

// Subscriber represents a subscriber with a channel and an optional filter.
type Subscriber struct {
	ch     chan types.Event
	filter types.FilterFunc
}

// NewEventBus creates a new instance of EventBus with a default retry policy and dead-letter queue.
func NewEventBus(obs *observability.Observability) *EventBus {
	cfg, err := config.LoadEventsConfig(obs.Logger)
	if err == nil {
		obs.Logger.Fatal("Failed to load resilience configuration", zap.Error(err))
	}

	return &EventBus{
		subscribers: make(map[string]map[int][]*Subscriber),
		shutdown:    make(chan struct{}),
		defaultRetryPolicy: types.RetryPolicy{
			MaxRetries: cfg.EventBus.DefaultRetryPolicy.MaxRetries,
			Backoff:    cfg.EventBus.DefaultRetryPolicy.Backoff,
		},
		bulkheads:        make(map[string]chan types.Event),
		bulkheadCapacity: cfg.EventBus.BulkheadCapacity,
		dlq: &types.DeadLetterQueue{
			Queue:   make(chan types.Event, cfg.EventBus.DLQ.MaxSize),
			MaxSize: cfg.EventBus.DLQ.MaxSize,
		},
		observability: obs,
	}
}

// Publish sends an event to all subscribers of that event type, with automatic retry.
func (eb *EventBus) Publish(event types.Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	eb.observability.Logger.Info("Publishing event", zap.String("event_name", event.Name), zap.Int("priority", event.Priority))
	metrics.IncrementEventPublished(event.Name)

	// Use bulkhead to prevent overloading
	bulkhead, exists := eb.bulkheads[event.Name]
	if !exists {
		bulkhead = make(chan types.Event, eb.bulkheadCapacity)
		eb.bulkheads[event.Name] = bulkhead
		go eb.processBulkhead(event.Name, bulkhead)
	}

	select {
	case bulkhead <- event:
		eb.observability.Logger.Info("Event added to bulkhead", zap.String("event_name", event.Name))
	default:
		eb.observability.Logger.Warn("Bulkhead full, dropping event", zap.String("event_name", event.Name))
		metrics.IncrementEventDropped(event.Name)
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

	eb.observability.Logger.Info("Subscriber added", zap.String("event_name", eventName), zap.Int("priority", priority))
	metrics.IncrementSubscriberAdded(eventName)
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
					eb.observability.Logger.Info("Subscriber removed", zap.String("event_name", eventName), zap.Int("priority", priority))
					metrics.IncrementSubscriberRemoved(eventName)
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
	eb.observability.Logger.Info("EventBus has been shut down gracefully")
}

// processBulkhead processes events from a bulkhead channel.
func (eb *EventBus) processBulkhead(eventName string, bulkhead chan types.Event) {
	for {
		select {
		case <-eb.shutdown:
			eb.observability.Logger.Info("Shutting down bulkhead processor", zap.String("event_name", eventName))
			return
		case event := <-bulkhead:
			metrics.IncrementEventProcessed(event.Name)
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
			eb.observability.Logger.Info("Event sent successfully", zap.String("event_name", event.Name), zap.Int("attempt", i+1))
			metrics.IncrementEventSent(event.Name)
			return true
		default:
			eb.observability.Logger.Warn("Subscriber channel full, retrying event", zap.String("event_name", event.Name), zap.Int("attempt", i+1), zap.Int("max_retries", eb.defaultRetryPolicy.MaxRetries))
			time.Sleep(eb.defaultRetryPolicy.Backoff)
		}
	}
	eb.observability.Logger.Error("Failed to send event after max attempts", zap.String("event_name", event.Name), zap.Int("max_retries", eb.defaultRetryPolicy.MaxRetries))
	metrics.IncrementEventFailed(event.Name)
	return false
}

// sendToDLQ sends an event to the Dead-Letter Queue for further investigation.
func (eb *EventBus) sendToDLQ(event types.Event) {
	select {
	case eb.dlq.Queue <- event:
		eb.observability.Logger.Warn("Event sent to Dead-Letter Queue", zap.String("event_name", event.Name))
		metrics.IncrementEventDQL(event.Name)
	default:
		eb.observability.Logger.Error("Dead-Letter Queue full, dropping event", zap.String("event_name", event.Name))
		metrics.IncrementEventDropped(event.Name)
	}
}
