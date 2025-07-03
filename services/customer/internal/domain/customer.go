package domain

import (
	"regexp"
	"time"

	"github.com/google/uuid"
)

// CustomerTier represents customer tier levels
type CustomerTier string

const (
	CustomerTierBronze   CustomerTier = "bronze"
	CustomerTierSilver   CustomerTier = "silver"
	CustomerTierGold     CustomerTier = "gold"
	CustomerTierPlatinum CustomerTier = "platinum"
	CustomerTierDiamond  CustomerTier = "diamond"

	// Maintain old constants for backward compatibility
	TierBronze   = CustomerTierBronze
	TierSilver   = CustomerTierSilver
	TierGold     = CustomerTierGold
	TierPlatinum = CustomerTierPlatinum
	TierDiamond  = CustomerTierDiamond
)

// AddressType represents address types
type AddressType string

const (
	AddressTypeHome     AddressType = "home"
	AddressTypeWork     AddressType = "work"
	AddressTypeBilling  AddressType = "billing"
	AddressTypeShipping AddressType = "shipping"
)

// Email regex pattern
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Phone regex pattern for Thai phone numbers
var phoneRegex = regexp.MustCompile(`^(\+66|0)[0-9]{8,9}$`)

// NewCustomer creates a new customer with validation
func NewCustomer(email, phone, firstName, lastName string) (*Customer, error) {
	if firstName == "" {
		return nil, ErrInvalidFirstName
	}
	if lastName == "" {
		return nil, ErrInvalidLastName
	}
	if email == "" {
		return nil, ErrInvalidEmail
	}
	if !emailRegex.MatchString(email) {
		return nil, ErrInvalidEmail
	}
	if phone == "" {
		return nil, ErrInvalidPhone
	}
	if !phoneRegex.MatchString(phone) {
		return nil, ErrInvalidPhone
	}

	now := time.Now()
	customer := &Customer{
		ID:        uuid.New(),
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Phone:     phone,
		Tier:      CustomerTierBronze,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return customer, nil
}

// UpdateProfile updates customer profile information
func (c *Customer) UpdateProfile(email, phone, firstName, lastName string) error {
	if firstName == "" {
		return ErrInvalidFirstName
	}
	if lastName == "" {
		return ErrInvalidLastName
	}
	if email == "" {
		return ErrInvalidEmail
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	if phone == "" {
		return ErrInvalidPhone
	}
	if !phoneRegex.MatchString(phone) {
		return ErrInvalidPhone
	}

	c.FirstName = firstName
	c.LastName = lastName
	c.Email = email
	c.Phone = phone
	c.UpdatedAt = time.Now()

	return nil
}

// UpdateTier updates customer tier
func (c *Customer) UpdateTier(tier CustomerTier) {
	c.Tier = tier
	c.UpdatedAt = time.Now()
}

// SetLoyverseID sets the Loyverse ID
func (c *Customer) SetLoyverseID(loyverseID string) {
	c.LoyverseID = &loyverseID
	c.UpdatedAt = time.Now()
}

// Deactivate deactivates the customer
func (c *Customer) Deactivate() {
	c.IsActive = false
	c.UpdatedAt = time.Now()
}

// Activate activates the customer
func (c *Customer) Activate() {
	c.IsActive = true
	c.UpdatedAt = time.Now()
}

// Customer represents a customer entity
type Customer struct {
	ID              uuid.UUID    `json:"id" db:"id"`
	FirstName       string       `json:"first_name" db:"first_name"`
	LastName        string       `json:"last_name" db:"last_name"`
	Email           string       `json:"email" db:"email"`
	Phone           string       `json:"phone" db:"phone"`
	DateOfBirth     *time.Time   `json:"date_of_birth" db:"date_of_birth"`
	Gender          *string      `json:"gender" db:"gender"`
	Tier            CustomerTier `json:"tier" db:"tier"`
	LoyverseID      *string      `json:"loyverse_id" db:"loyverse_id"`
	TotalSpent      float64      `json:"total_spent" db:"total_spent"`
	OrderCount      int          `json:"order_count" db:"order_count"`
	LastOrderDate   *time.Time   `json:"last_order_date" db:"last_order_date"`
	DeliveryRouteID *uuid.UUID   `json:"delivery_route_id" db:"delivery_route_id"`
	IsActive        bool         `json:"is_active" db:"is_active"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at" db:"updated_at"`
}

// CustomerAddress represents a customer address
type CustomerAddress struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	CustomerID     uuid.UUID  `json:"customer_id" db:"customer_id"`
	Type           string     `json:"type" db:"type"` // home, work, billing, shipping
	Label          string     `json:"label" db:"label"`
	AddressLine1   string     `json:"address_line1" db:"address_line1"`
	AddressLine2   *string    `json:"address_line2" db:"address_line2"`
	SubDistrict    string     `json:"sub_district" db:"sub_district"`
	District       string     `json:"district" db:"district"`
	Province       string     `json:"province" db:"province"`
	ThaiAddressID  *uuid.UUID `json:"thai_address_id" db:"thai_address_id"`
	PostalCode     string     `json:"postal_code" db:"postal_code"`
	Latitude       *float64   `json:"latitude" db:"latitude"`
	Longitude      *float64   `json:"longitude" db:"longitude"`
	IsDefault      bool       `json:"is_default" db:"is_default"`
	DeliveryNotes  *string    `json:"delivery_notes" db:"delivery_notes"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// ThaiAddress represents Thai administrative divisions
type ThaiAddress struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Province     string    `json:"province" db:"province"`
	District     string    `json:"district" db:"district"`
	Subdistrict  string    `json:"subdistrict" db:"subdistrict"`
	PostalCode   string    `json:"postal_code" db:"postal_code"`
	ProvinceCode string    `json:"province_code" db:"province_code"`
	DistrictCode string    `json:"district_code" db:"district_code"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// DeliveryRoute represents delivery routes
type DeliveryRoute struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CustomerWithAddresses includes customer with their addresses
type CustomerWithAddresses struct {
	Customer  Customer          `json:"customer"`
	Addresses []CustomerAddress `json:"addresses"`
}

// ValidateCustomer validates customer data
func (c *Customer) ValidateCustomer() error {
	if c.FirstName == "" {
		return ErrInvalidFirstName
	}
	if c.LastName == "" {
		return ErrInvalidLastName
	}
	if c.Email == "" {
		return ErrInvalidEmail
	}
	if c.Phone == "" {
		return ErrInvalidPhone
	}
	return nil
}

// GetFullName returns customer's full name
func (c *Customer) GetFullName() string {
	return c.FirstName + " " + c.LastName
}

// CalculateTier calculates customer tier based on total spent
func (c *Customer) CalculateTier() CustomerTier {
	switch {
	case c.TotalSpent >= 100000: // 100k THB
		return TierDiamond
	case c.TotalSpent >= 50000: // 50k THB
		return TierPlatinum
	case c.TotalSpent >= 20000: // 20k THB
		return TierGold
	case c.TotalSpent >= 5000: // 5k THB
		return TierSilver
	default:
		return TierBronze
	}
}

// UpdateTotalSpent updates customer's total spent and recalculates tier
func (c *Customer) UpdateTotalSpent(amount float64) {
	c.TotalSpent += amount
	c.Tier = c.CalculateTier()
	c.OrderCount++
	now := time.Now()
	c.LastOrderDate = &now
	c.UpdatedAt = now
}

// NewCustomerAddress creates a new customer address with validation
func NewCustomerAddress(customerID uuid.UUID, addressType AddressType, label, addressLine1, addressLine2, subDistrict, district, province, postalCode string, isDefault bool) (*CustomerAddress, error) {
	if addressLine1 == "" {
		return nil, ErrInvalidAddressLine1
	}
	if subDistrict == "" {
		return nil, ErrInvalidSubDistrict
	}
	if district == "" {
		return nil, ErrInvalidDistrict
	}
	if province == "" {
		return nil, ErrInvalidProvince
	}
	if postalCode == "" {
		return nil, ErrInvalidPostalCode
	}

	now := time.Now()
	address := &CustomerAddress{
		ID:           uuid.New(),
		CustomerID:   customerID,
		Type:         string(addressType),
		Label:        label,
		AddressLine1: addressLine1,
		AddressLine2: &addressLine2,
		SubDistrict:  subDistrict,
		District:     district,
		Province:     province,
		PostalCode:   postalCode,
		IsDefault:    isDefault,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if addressLine2 == "" {
		address.AddressLine2 = nil
	}

	return address, nil
}

// Update updates address information
func (a *CustomerAddress) Update(addressType AddressType, label, addressLine1, addressLine2, subDistrict, district, province, postalCode string, isDefault bool) error {
	if addressLine1 == "" {
		return ErrInvalidAddressLine1
	}
	if subDistrict == "" {
		return ErrInvalidSubDistrict
	}
	if district == "" {
		return ErrInvalidDistrict
	}
	if province == "" {
		return ErrInvalidProvince
	}
	if postalCode == "" {
		return ErrInvalidPostalCode
	}

	a.Type = string(addressType)
	a.Label = label
	a.AddressLine1 = addressLine1
	a.AddressLine2 = &addressLine2
	a.SubDistrict = subDistrict
	a.District = district
	a.Province = province
	a.PostalCode = postalCode
	a.IsDefault = isDefault
	a.UpdatedAt = time.Now()

	if addressLine2 == "" {
		a.AddressLine2 = nil
	}

	return nil
}

// SetAsDefault sets address as default
func (a *CustomerAddress) SetAsDefault() {
	a.IsDefault = true
	a.UpdatedAt = time.Now()
}

// UnsetAsDefault unsets address as default
func (a *CustomerAddress) UnsetAsDefault() {
	a.IsDefault = false
	a.UpdatedAt = time.Now()
}

// AddressSuggestion represents address suggestion response
type AddressSuggestion struct {
	ID            string `json:"id"`
	DisplayText   string `json:"display_text"` // "หัวหมาก > บางกะปิ > กรุงเทพมหานคร (10240)"
	Subdistrict   string `json:"subdistrict"`
	District      string `json:"district"`
	Province      string `json:"province"`
	PostalCode    string `json:"postal_code"`
	DeliveryRoute string `json:"delivery_route,omitempty"`
}
