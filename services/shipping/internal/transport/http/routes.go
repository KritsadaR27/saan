package http

import (
	"github.com/gorilla/mux"
	"shipping/internal/transport/http/handler"
	"shipping/internal/transport/http/middleware"
)

// setupRoutes configures all HTTP routes for the shipping service
func setupRoutes(
	router *mux.Router,
	deliveryHandler *handler.DeliveryHandler,
	vehicleHandler *handler.VehicleHandler,
	providerHandler *handler.ProviderHandler,
	routingHandler *handler.RoutingHandler,
	trackingHandler *handler.TrackingHandler,
	coverageHandler *handler.CoverageHandler,
	healthHandler *handler.HealthHandler,
) {
	// Add middleware
	router.Use(middleware.Logger())
	router.Use(middleware.CORS())
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestID())

	// Health check endpoint
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")
	router.HandleFunc("/ready", healthHandler.Ready).Methods("GET")

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Delivery routes
	deliveryRoutes := api.PathPrefix("/deliveries").Subrouter()
	deliveryRoutes.HandleFunc("", deliveryHandler.CreateDelivery).Methods("POST")
	deliveryRoutes.HandleFunc("", deliveryHandler.GetDeliveries).Methods("GET")
	deliveryRoutes.HandleFunc("/{id}", deliveryHandler.GetDelivery).Methods("GET")
	deliveryRoutes.HandleFunc("/{id}", deliveryHandler.UpdateDelivery).Methods("PUT")
	deliveryRoutes.HandleFunc("/{id}", deliveryHandler.DeleteDelivery).Methods("DELETE")
	deliveryRoutes.HandleFunc("/{id}/status", deliveryHandler.UpdateDeliveryStatus).Methods("PATCH")
	deliveryRoutes.HandleFunc("/{id}/tracking", deliveryHandler.GetDeliveryTracking).Methods("GET")
	deliveryRoutes.HandleFunc("/tracking/{tracking_number}", deliveryHandler.GetDeliveryByTracking).Methods("GET")
	deliveryRoutes.HandleFunc("/search", deliveryHandler.SearchDeliveries).Methods("GET")

	// Vehicle routes
	vehicleRoutes := api.PathPrefix("/vehicles").Subrouter()
	vehicleRoutes.HandleFunc("", vehicleHandler.GetVehicles).Methods("GET")
	vehicleRoutes.HandleFunc("/{id}", vehicleHandler.GetVehicle).Methods("GET")
	vehicleRoutes.HandleFunc("", vehicleHandler.CreateVehicle).Methods("POST")
	vehicleRoutes.HandleFunc("/{id}", vehicleHandler.UpdateVehicle).Methods("PUT")
	vehicleRoutes.HandleFunc("/{id}", vehicleHandler.DeleteVehicle).Methods("DELETE")

	// Provider routes
	providerRoutes := api.PathPrefix("/providers").Subrouter()
	providerRoutes.HandleFunc("", providerHandler.GetProviders).Methods("GET")
	providerRoutes.HandleFunc("/{id}", providerHandler.GetProvider).Methods("GET")
	providerRoutes.HandleFunc("", providerHandler.CreateProvider).Methods("POST")
	providerRoutes.HandleFunc("/{id}", providerHandler.UpdateProvider).Methods("PUT")
	providerRoutes.HandleFunc("/{id}", providerHandler.DeleteProvider).Methods("DELETE")

	// Routing routes
	routingRoutes := api.PathPrefix("/routes").Subrouter()
	routingRoutes.HandleFunc("/calculate", routingHandler.CalculateRoute).Methods("POST")
	routingRoutes.HandleFunc("/optimize", routingHandler.OptimizeRoutes).Methods("POST")
	routingRoutes.HandleFunc("/{id}", routingHandler.GetRoute).Methods("GET")

	// Tracking routes
	trackingRoutes := api.PathPrefix("/tracking").Subrouter()
	trackingRoutes.HandleFunc("/{tracking_number}", trackingHandler.GetTracking).Methods("GET")
	trackingRoutes.HandleFunc("/{tracking_number}/update", trackingHandler.UpdateTracking).Methods("POST")

	// Coverage routes
	coverageRoutes := api.PathPrefix("/coverage").Subrouter()
	coverageRoutes.HandleFunc("/areas", coverageHandler.GetCoverageAreas).Methods("GET")
	coverageRoutes.HandleFunc("/check", coverageHandler.CheckCoverage).Methods("POST")
}
