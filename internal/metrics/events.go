package metrics

import (
	"github.com/goletan/observability-library/pkg"
	"github.com/prometheus/client_golang/prometheus"
)

type EventsMetrics struct{}

// Event Metrics: Track event processing metrics
var (
	EventPublished = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goletan",
			Subsystem: "events_service",
			Name:      "published_total",
			Help:      "Total number of events published.",
		},
		[]string{"event_type"},
	)
	EventDropped = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goletan",
			Subsystem: "events_service",
			Name:      "dropped_total",
			Help:      "Total number of events dropped.",
		},
		[]string{"event_type"},
	)
	EventProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goletan",
			Subsystem: "events_service",
			Name:      "processed_total",
			Help:      "Total number of events processed.",
		},
		[]string{"event_type"},
	)
	EventSent = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goletan",
			Subsystem: "events_service",
			Name:      "sent_total",
			Help:      "Total number of events sent.",
		},
		[]string{"event_type"},
	)
	EventFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goletan",
			Subsystem: "events_service",
			Name:      "failed_total",
			Help:      "Total number of events failed.",
		},
		[]string{"event_type"},
	)
	EventDLQ = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "goletan",
			Subsystem: "events_service",
			Name:      "dlq_total",
			Help:      "Total number of events dead letter queue.",
		},
		[]string{"event_type"},
	)
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
	EventProcessingLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "goletan",
			Subsystem: "events_service",
			Name:      "processing_latency_seconds",
			Help:      "Histogram of event processing latencies.",
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~16s
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

	if err := prometheus.Register(EventPublished); err != nil {
		return err
	}

	if err := prometheus.Register(EventDropped); err != nil {
		return err
	}

	if err := prometheus.Register(EventProcessed); err != nil {
		return err
	}

	if err := prometheus.Register(EventSent); err != nil {
		return err
	}

	if err := prometheus.Register(EventFailed); err != nil {
		return err
	}

	if err := prometheus.Register(EventDLQ); err != nil {
		return err
	}

	if err := prometheus.Register(SubscriberAdded); err != nil {
		return err
	}

	if err := prometheus.Register(SubscriberRemoved); err != nil {
		return err
	}

	if err := prometheus.Register(EventProcessingLatency); err != nil {
		return err
	}

	if err := prometheus.Register(AIDecisions); err != nil {
		return err
	}

	return nil
}

// IncrementEventPublished increments the event published counter.
func IncrementEventPublished(eventType string) {
	EventPublished.WithLabelValues(eventType).Inc()
}

// IncrementEventDropped increments the event dropped counter.
func IncrementEventDropped(eventType string) {
	EventDropped.WithLabelValues(eventType).Inc()
}

// IncrementEventProcessed increments the event processed counter.
func IncrementEventProcessed(eventType string) {
	EventProcessed.WithLabelValues(eventType).Inc()
}

// IncrementEventSent increments the event sent counter.
func IncrementEventSent(eventType string) {
	EventSent.WithLabelValues(eventType).Inc()
}

// IncrementEventFailed increments the event failed counter.
func IncrementEventFailed(eventType string) {
	EventFailed.WithLabelValues(eventType).Inc()
}

// IncrementEventDQL increments the event Dead Letter Queue counter.
func IncrementEventDQL(eventType string) {
	EventDLQ.WithLabelValues(eventType).Inc()
}

// IncrementSubscriberAdded increments the number of subscriber added counter.
func IncrementSubscriberAdded(eventType string) {
	SubscriberAdded.WithLabelValues(eventType).Inc()
}

// IncrementSubscriberRemoved increments the number of subscriber removed counter.
func IncrementSubscriberRemoved(eventType string) {
	SubscriberRemoved.WithLabelValues(eventType).Inc()
}

// RecordProcessingLatency observes the processing latency for a specific event type.
func RecordProcessingLatency(eventType string, latencySeconds float64) {
	EventProcessingLatency.WithLabelValues(eventType).Observe(latencySeconds)
}
