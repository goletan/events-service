package metrics

import (
	"github.com/goletan/observability-library/pkg"
	"github.com/prometheus/client_golang/prometheus"
)

type EventsMetrics struct{}

// Event Metrics: Track event processing metrics
var (
	SubscriberAdded = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goletan",
			Subsystem: "events_service",
			Name:      "subscriber_added_total",
			Help:      "Total number of subscriber added.",
		},
		[]string{"event_type"},
	)
	SubscriberRemoved = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goletan",
			Subsystem: "events_service",
			Name:      "subscriber_removed_total",
			Help:      "Total number of subscriber removed.",
		},
		[]string{"event_type"},
	)
	AIDecisions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goletan",
			Subsystem: "events_service",
			Name:      "ai_decisions_total",
			Help:      "Total number of AI-driven decisions in event processing.",
		},
		[]string{"decision_type"},
	)
)

func InitMetrics(observer *observability.Observability) {
	observer.Metrics.Register(&EventsMetrics{})
}

func (em *EventsMetrics) Register() error {

	if err := prometheus.Register(SubscriberAdded); err != nil {
		return err
	}

	if err := prometheus.Register(SubscriberRemoved); err != nil {
		return err
	}

	if err := prometheus.Register(AIDecisions); err != nil {
		return err
	}

	return nil
}

// IncrementSubscriberAdded increments the number of subscriber added counter.
func IncrementSubscriberAdded(eventType string) {
	SubscriberAdded.WithLabelValues(eventType).Inc()
}

// IncrementSubscriberRemoved increments the number of subscriber removed counter.
func IncrementSubscriberRemoved(eventType string) {
	SubscriberRemoved.WithLabelValues(eventType).Inc()
}
