package entity

import "errors"

// Customer domain errors
var (
	ErrCustomerNotFound      = errors.New("customer not found")
	ErrCustomerExists        = errors.New("customer already exists")
	ErrCustomerAlreadyExists = errors.New("customer already exists") // Alias for consistency
	ErrCustomerEmailExists   = errors.New("customer with this email already exists")
	ErrCustomerPhoneExists   = errors.New("customer with this phone already exists")
	ErrInvalidCustomerID     = errors.New("invalid customer ID")
	ErrInvalidFirstName      = errors.New("first name is required")
	ErrInvalidLastName       = errors.New("last name is required")
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrInvalidPhone          = errors.New("invalid phone format")
	ErrInvalidTier           = errors.New("invalid customer tier")
)

// Address domain errors
var (
	ErrAddressNotFound     = errors.New("address not found")
	ErrInvalidAddressType  = errors.New("invalid address type")
	ErrInvalidAddressLine1 = errors.New("address line 1 is required")
	ErrInvalidPostalCode   = errors.New("postal code is required")
	ErrInvalidAddressData  = errors.New("invalid address data")
)

// Thai address domain errors
var (
	ErrThaiAddressNotFound = errors.New("thai address not found")
	ErrInvalidProvince     = errors.New("province is required")
	ErrInvalidDistrict     = errors.New("district is required")
	ErrInvalidSubDistrict  = errors.New("sub district is required")
	ErrInvalidSubdistrict  = errors.New("subdistrict is required") // Keep for backward compatibility
)

// Delivery route domain errors
var (
	ErrDeliveryRouteNotFound = errors.New("delivery route not found")
	ErrInvalidRouteName      = errors.New("route name is required")
)

// Loyverse integration errors
var (
	ErrLoyverseCustomerNotFound = errors.New("loyverse customer not found")
	ErrLoyverseSyncFailed       = errors.New("loyverse sync failed")
	ErrLoyverseAPIError         = errors.New("loyverse API error")
)

// Points system errors
var (
	ErrInsufficientPoints     = errors.New("insufficient points balance")
	ErrInvalidPointsAmount    = errors.New("invalid points amount")
	ErrPointsTransactionFailed = errors.New("points transaction failed")
)

// VIP tier errors
var (
	ErrInvalidTierLevel       = errors.New("invalid tier level")
	ErrTierUpgradeFailed      = errors.New("tier upgrade failed")
	ErrTierBenefitsNotFound   = errors.New("tier benefits not found")
	ErrVIPTierNotFound        = errors.New("VIP tier not found")
)

// LINE integration errors
var (
	ErrLINEUserNotFound       = errors.New("LINE user not found")
	ErrLINEIntegrationFailed  = errors.New("LINE integration failed")
	ErrDigitalCardNotIssued   = errors.New("digital card not issued")
)

// Customer code errors
var (
	ErrCustomerCodeExists     = errors.New("customer code already exists")
	ErrInvalidCustomerCode    = errors.New("invalid customer code format")
)
