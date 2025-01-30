package metrics

import (
	"github.com/goletan/observability-library/pkg"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type EventsMetrics struct {
	obs *observability.Observability

	eventDQL       *prometheus.CounterVec
	eventPublished *prometheus.CounterVec
	eventDropped   *prometheus.CounterVec
	eventProcessed *prometheus.CounterVec
	eventSent      *prometheus.CounterVec
	eventFailed    *prometheus.CounterVec

	subscriberAdded   *prometheus.CounterVec
	subscriberRemoved *prometheus.CounterVec

	publishedTotal      *prometheus.CounterVec
	failedTotal         *prometheus.CounterVec
	latency             *prometheus.HistogramVec
	serializationErrors *prometheus.CounterVec
	aiDecisions         *prometheus.CounterVec
}

func InitMetrics(obs *observability.Observability) *EventsMetrics {
	return &EventsMetrics{
		obs: obs,
		eventDQL: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "goletan",
				Subsystem: "events_service",
				Name:      "dlq_total",
				Help:      "Total number of events dead letter queue.",
			},
			[]string{"event_type"},
		),
		eventPublished: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "goletan",
				Subsystem: "events_service",
				Name:      "published_total",
				Help:      "Total number of events published.",
			},
			[]string{"event_type"},
		),
		eventDropped: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "goletan",
				Subsystem: "events_service",
				Name:      "dropped_total",
				Help:      "Total number of events dropped.",
			},
			[]string{"event_type"},
		),
		eventProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "goletan",
				Subsystem: "events_service",
				Name:      "processed_total",
				Help:      "Total number of events processed.",
			},
			[]string{"event_type"},
		),
		eventSent: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "goletan",
				Subsystem: "events_service",
				Name:      "sent_total",
				Help:      "Total number of events sent.",
			},
			[]string{"event_type"},
		),
		eventFailed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "goletan",
				Subsystem: "events_service",
				Name:      "failed_total",
				Help:      "Total number of events failed.",
			},
			[]string{"event_type"},
		),
		subscriberAdded: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "goletan",
				Subsystem: "events_service",
				Name:      "subscriber_added_total",
				Help:      "Total number of subscriber added.",
			},
			[]string{"event_type"},
		),
		subscriberRemoved: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "goletan",
				Subsystem: "events_service",
				Name:      "subscriber_removed_total",
				Help:      "Total number of subscriber removed.",
			},
			[]string{"event_type"},
		),
		publishedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "goletan",
				Subsystem: "events_library",
				Name:      "published_total",
				Help:      "Total number of events successfully published.",
			},
			[]string{"event_type"},
		),
		failedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "goletan",
				Subsystem: "events_library",
				Name:      "failed_total",
				Help:      "Total number of events failed to publish.",
			},
			[]string{"event_type"},
		),
		latency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "goletan",
				Subsystem: "events_library",
				Name:      "processing_latency_seconds",
				Help:      "Time taken to process events.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"event_type"},
		),
		serializationErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "goletan",
				Subsystem: "events_library",
				Name:      "serialization_errors_total",
				Help:      "Total number of serialization errors.",
			},
			[]string{"operation"},
		),
		aiDecisions: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "goletan",
				Subsystem: "events_service",
				Name:      "ai_decisions_total",
				Help:      "Total number of AI-driven decisions in event processing.",
			},
			[]string{"decision_type"},
		),
	}
}

func (m *EventsMetrics) Register() {

	prometheus.MustRegister(m.eventDQL)
	prometheus.MustRegister(m.eventFailed)
	prometheus.MustRegister(m.eventDropped)
	prometheus.MustRegister(m.eventPublished)
	prometheus.MustRegister(m.eventProcessed)
	prometheus.MustRegister(m.eventSent)

	prometheus.MustRegister(m.subscriberAdded)
	prometheus.MustRegister(m.subscriberRemoved)

	prometheus.MustRegister(m.publishedTotal)
	prometheus.MustRegister(m.failedTotal)
	prometheus.MustRegister(m.latency)
	prometheus.MustRegister(m.serializationErrors)
	prometheus.MustRegister(m.aiDecisions)

	m.obs.Logger.Info("Events metrics registered successfully")
}

func (m *EventsMetrics) IncrementPublished(eventType string) {
	m.publishedTotal.WithLabelValues(eventType).Inc()
	m.obs.Logger.Info("Incremented published events counter", zap.String("event_type", eventType))
}

func (m *EventsMetrics) IncrementFailed(eventType string) {
	m.failedTotal.WithLabelValues(eventType).Inc()
	m.obs.Logger.Info("Incremented failed events counter", zap.String("event_type", eventType))
}

func (m *EventsMetrics) ObserveLatency(eventType string, duration float64) {
	m.latency.WithLabelValues(eventType).Observe(duration)
	m.obs.Logger.Info("Recorded event processing latency", zap.String("event_type", eventType), zap.Float64("latency", duration))
}

func (m *EventsMetrics) IncrementSerializationErrors(eventType string) {
	m.serializationErrors.WithLabelValues(eventType).Inc()
	m.obs.Logger.Info("Incremented serialization error counter", zap.String("event_type", eventType))
}

// IncrementEventDQL increments the event Dead Letter Queue counter.
func (m *EventsMetrics) IncrementEventDQL(eventType string) {
	m.eventDQL.WithLabelValues(eventType).Inc()
	m.obs.Logger.Info("Incremented Event DLQ error counter", zap.String("event_type", eventType))
}

// IncrementEventFailed increments the event failed counter.
func (m *EventsMetrics) IncrementEventFailed(eventType string) {
	m.eventFailed.WithLabelValues(eventType).Inc()
	m.obs.Logger.Info("Incremented Event Failed error counter", zap.String("event_type", eventType))
}

// IncrementEventPublished increments the event published counter.
func (m *EventsMetrics) IncrementEventPublished(eventType string) {
	m.eventPublished.WithLabelValues(eventType).Inc()
	m.obs.Logger.Info("Incremented Event Published error counter", zap.String("event_type", eventType))
}

// IncrementEventDropped increments the event dropped counter.
func (m *EventsMetrics) IncrementEventDropped(eventType string) {
	m.eventDropped.WithLabelValues(eventType).Inc()
	m.obs.Logger.Info("Incremented Event Dropped error counter", zap.String("event_type", eventType))
}

// IncrementEventProcessed increments the event processed counter.
func (m *EventsMetrics) IncrementEventProcessed(eventType string) {
	m.eventProcessed.WithLabelValues(eventType).Inc()
	m.obs.Logger.Info("Incremented Event Processed error counter", zap.String("event_type", eventType))
}

// IncrementEventSent increments the event sent counter.
func (m *EventsMetrics) IncrementEventSent(eventType string) {
	m.eventSent.WithLabelValues(eventType).Inc()
	m.obs.Logger.Info("Incremented Event Sent error counter", zap.String("event_type", eventType))
}

// IncrementSubscriberAdded increments the number of subscriber added counter.
func (m *EventsMetrics) IncrementSubscriberAdded(eventType string) {
	m.subscriberAdded.WithLabelValues(eventType).Inc()
	m.obs.Logger.Info("Incremented Subscriber Added error counter", zap.String("event_type", eventType))
}

// IncrementSubscriberRemoved increments the number of subscriber removed counter.
func (m *EventsMetrics) IncrementSubscriberRemoved(eventType string) {
	m.subscriberRemoved.WithLabelValues(eventType).Inc()
	m.obs.Logger.Info("Incremented Subscriber Removed error counter", zap.String("event_type", eventType))
}
