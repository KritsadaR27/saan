package domain

import "errors"

var (
	// Order errors
	ErrOrderNotFound           = errors.New("order not found")
	ErrOrderAlreadyExists      = errors.New("order already exists")
	ErrOrderAlreadyCancelled   = errors.New("order already cancelled")
	ErrOrderCannotBeCancelled  = errors.New("order cannot be cancelled")
	ErrInvalidOrderData        = errors.New("invalid order data")
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrOrderCannotBeModified   = errors.New("order cannot be modified in current status")
	ErrInvalidOrderStatus      = errors.New("invalid order status for this operation")
	ErrUnauthorizedStockOverride = errors.New("unauthorized to perform stock override")
	
	// Order item errors
	ErrOrderItemNotFound     = errors.New("order item not found")
	ErrInvalidOrderItemData  = errors.New("invalid order item data")
	ErrInvalidQuantity       = errors.New("invalid quantity")
	ErrInvalidPrice          = errors.New("invalid price")
	ErrInvalidAmount         = errors.New("invalid amount")
	
	// Event errors
	ErrEventNotFound = errors.New("event not found")
	
	// Customer errors
	ErrInvalidCustomerID = errors.New("invalid customer ID")
	
	// Repository errors
	ErrDatabaseConnection = errors.New("database connection error")
	ErrDatabaseOperation  = errors.New("database operation error")
)
