package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"shipping/internal/application"
	"shipping/internal/transport/http/handler"
)

// Server represents the HTTP server
type Server struct {
	server          *http.Server
	deliveryHandler *handler.DeliveryHandler
	vehicleHandler  *handler.VehicleHandler
	providerHandler *handler.ProviderHandler
	routingHandler  *handler.RoutingHandler
	trackingHandler *handler.TrackingHandler
	coverageHandler *handler.CoverageHandler
	healthHandler   *handler.HealthHandler
}

// NewServer creates a new HTTP server
func NewServer(
	port string,
	deliveryUseCase *application.DeliveryUsecase,
	vehicleUseCase *application.VehicleUseCase,
	providerUseCase *application.ProviderUseCase,
	routingUseCase *application.RoutingUseCase,
	trackingUseCase *application.TrackingUseCase,
	coverageUseCase *application.CoverageUseCase,
) *Server {
	// Create handlers
	deliveryHandler := handler.NewDeliveryHandler(deliveryUseCase)
	vehicleHandler := handler.NewVehicleHandler(vehicleUseCase)
	providerHandler := handler.NewProviderHandler(providerUseCase)
	routingHandler := handler.NewRoutingHandler(routingUseCase)
	trackingHandler := handler.NewTrackingHandler(trackingUseCase)
	coverageHandler := handler.NewCoverageHandler(coverageUseCase)
	healthHandler := handler.NewHealthHandler()

	// Create router and setup routes
	router := mux.NewRouter()
	setupRoutes(router, deliveryHandler, vehicleHandler, providerHandler, 
		routingHandler, trackingHandler, coverageHandler, healthHandler)
	
	// Create server instance
	server := &Server{
		deliveryHandler: deliveryHandler,
		vehicleHandler:  vehicleHandler,
		providerHandler: providerHandler,
		routingHandler:  routingHandler,
		trackingHandler: trackingHandler,
		coverageHandler: coverageHandler,
		healthHandler:   healthHandler,
	}
	
	// Create HTTP server
	server.server = &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server
}

// setupRoutes configures all API routes
// Start starts the HTTP server
func (s *Server) Start() error {
	fmt.Printf("Starting HTTP server on port %s\n", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("Stopping HTTP server...")
	return s.server.Shutdown(ctx)
}
