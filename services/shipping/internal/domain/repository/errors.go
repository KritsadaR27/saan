package repository

import "errors"

// Repository errors
var (
	// Common errors
	ErrNotFound            = errors.New("resource not found")
	ErrDuplicateKey        = errors.New("duplicate key constraint")
	ErrInvalidInput        = errors.New("invalid input data")
	ErrConnectionFailed    = errors.New("database connection failed")
	
	// Delivery errors
	ErrDeliveryNotFound    = errors.New("delivery not found")
	ErrInvalidDeliveryData = errors.New("invalid delivery data")
	
	// Vehicle errors
	ErrVehicleNotFound     = errors.New("vehicle not found")
	ErrVehicleInUse        = errors.New("vehicle is currently in use")
	
	// Route errors
	ErrRouteNotFound       = errors.New("route not found")
	ErrRouteOptimization   = errors.New("route optimization failed")
	
	// Provider errors
	ErrProviderNotFound    = errors.New("provider not found")
	ErrProviderInactive    = errors.New("provider is inactive")
	
	// Snapshot errors
	ErrSnapshotNotFound    = errors.New("snapshot not found")
	ErrInvalidSnapshotData = errors.New("invalid snapshot data")
	
	// Manual task errors
	ErrManualTaskNotFound  = errors.New("manual task not found")
	ErrTaskAlreadyAssigned = errors.New("task is already assigned")
	ErrTaskNotAssigned     = errors.New("task is not assigned")
	
	// Coverage area errors
	ErrCoverageAreaNotFound = errors.New("coverage area not found")
	ErrLocationNotCovered   = errors.New("location is not covered")
)
