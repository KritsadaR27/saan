package entity

import "errors"

// Domain errors for Payment Service
var (
	// Payment Transaction errors
	ErrInvalidPaymentTransactionID = errors.New("invalid payment transaction ID")
	ErrInvalidOrderID              = errors.New("invalid order ID")
	ErrInvalidCustomerID           = errors.New("invalid customer ID")
	ErrInvalidAmount               = errors.New("invalid amount: must be greater than 0")
	ErrInvalidCurrency             = errors.New("invalid currency: must be 3 characters")
	ErrInvalidPaymentMethod        = errors.New("invalid payment method")
	ErrInvalidPaymentChannel       = errors.New("invalid payment channel")
	ErrInvalidPaymentTiming        = errors.New("invalid payment timing")
	ErrInvalidPaymentStatus        = errors.New("invalid payment status")
	ErrPaymentAlreadyCompleted     = errors.New("payment already completed")
	ErrPaymentAlreadyCancelled     = errors.New("payment already cancelled")
	ErrPaymentNotFound             = errors.New("payment not found")

	// Loyverse Store errors
	ErrInvalidStoreID              = errors.New("invalid store ID")
	ErrInvalidStoreName            = errors.New("invalid store name: cannot be empty")
	ErrInvalidStoreCode            = errors.New("invalid store code: cannot be empty")
	ErrInvalidStoreType            = errors.New("invalid store type")
	ErrInvalidManagerID            = errors.New("invalid manager ID")
	ErrStoreNotActive              = errors.New("store is not active")
	ErrStoreNotFound               = errors.New("store not found")
	ErrStoreCapacityExceeded       = errors.New("store capacity exceeded")
	ErrInvalidBusinessHours        = errors.New("invalid business hours")
	ErrLoyverseIntegrationFailed   = errors.New("loyverse integration failed")

	// Payment Delivery Context errors
	ErrInvalidDeliveryID           = errors.New("invalid delivery ID")
	ErrInvalidDriverID             = errors.New("invalid driver ID")
	ErrInvalidDeliveryAddress      = errors.New("invalid delivery address: cannot be empty")
	ErrInvalidDeliveryStatus       = errors.New("invalid delivery status")
	ErrInvalidCODAmount            = errors.New("invalid COD amount")
	ErrInvalidGPSCoordinates       = errors.New("invalid GPS coordinates")
	ErrDeliveryContextNotFound     = errors.New("delivery context not found")
	ErrDeliveryAlreadyCompleted    = errors.New("delivery already completed")

	// Business Logic errors
	ErrInsufficientPayment         = errors.New("insufficient payment amount")
	ErrOverpaymentNotAllowed       = errors.New("overpayment not allowed")
	ErrRefundAmountExceeded        = errors.New("refund amount exceeds paid amount")
	ErrCODCollectionFailed         = errors.New("COD collection failed")
	ErrStoreAssignmentFailed       = errors.New("store assignment failed")
	ErrEventPublishingFailed       = errors.New("event publishing failed")
	ErrConcurrentModification      = errors.New("concurrent modification detected")

	// Validation errors
	ErrRequiredFieldMissing        = errors.New("required field is missing")
	ErrInvalidFieldFormat          = errors.New("invalid field format")
	ErrFieldTooLong                = errors.New("field value too long")
	ErrFieldTooShort               = errors.New("field value too short")
	ErrInvalidDateRange            = errors.New("invalid date range")
	ErrInvalidFilterCombination    = errors.New("invalid filter combination")

	// External Service errors
	ErrOrderServiceUnavailable     = errors.New("order service unavailable")
	ErrCustomerServiceUnavailable  = errors.New("customer service unavailable")
	ErrShippingServiceUnavailable  = errors.New("shipping service unavailable")
	ErrLoyverseServiceUnavailable  = errors.New("loyverse service unavailable")
	ErrDatabaseConnectionFailed    = errors.New("database connection failed")
	ErrRedisConnectionFailed       = errors.New("redis connection failed")
	ErrKafkaPublishFailed          = errors.New("kafka publish failed")
)
