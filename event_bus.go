// rest_server.go
package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/goletan/observability/logger"
	"github.com/goletan/services"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// RESTServer is an enhanced HTTP server that implements the Service interface.
type RESTServer struct {
	server *http.Server
	name   string
}

// NewRESTServer creates a new instance of the RESTServer.
func NewRESTServer(cfg Config) services.Service {
	r := mux.NewRouter()

	// Define middlewares for observability
	r.Use(loggingMiddleware)
	r.Use(metricsMiddleware)

	// Define your REST endpoints here, e.g.:
	r.HandleFunc("/health", healthHandler).Methods("GET")

	return &RESTServer{
		server: &http.Server{
			Addr:              cfg.Address, // Load from config
			Handler:           r,
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
		},
		name: "REST Server",
	}
}

// Name returns the service name.
func (s *RESTServer) Name() string {
	return s.name
}

// Initialize performs any initialization tasks needed by the service.
func (s *RESTServer) Initialize() error {
	logger.Info("Initializing REST server", zap.String("service", s.name))
	return nil
}

// Start starts the REST server.
func (s *RESTServer) Start() error {
	go func() {
		logger.Info("Starting REST server", zap.String("address", s.server.Addr))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start REST server", zap.Error(err))
		}
	}()
	return nil
}

// Stop gracefully stops the REST server.
func (s *RESTServer) Stop() error {
	logger.Info("Stopping REST server", zap.String("service", s.name))
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// Middleware for logging requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Incoming request", zap.String("method", r.Method), zap.String("url", r.URL.String()))
		next.ServeHTTP(w, r)
	})
}

// Middleware for collecting metrics.
func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increment request counter, collect duration, etc.
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		// You would use the metrics library to collect metrics here
		logger.Info("Request processed",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Duration("duration", duration),
		)
	})
}

// Health handler for health checks.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
