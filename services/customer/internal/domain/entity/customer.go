package entity

import (
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// CustomerTier represents customer tier levels (1-5)
type CustomerTier int

const (
	TierBronze   CustomerTier = 1  // 🥉 Bronze   - เริ่มต้น
	TierSilver   CustomerTier = 2  // 🥈 Silver   - ซื้อ 10,000 บาท
	TierGold     CustomerTier = 3  // 🥇 Gold     - ซื้อ 50,000 บาท  
	TierPlatinum CustomerTier = 4  // 💎 Platinum - ซื้อ 100,000 บาท
	TierDiamond  CustomerTier = 5  // 💍 Diamond  - ซื้อ 250,000 บาท
)

// CustomerTierString returns string representation of tier
func (t CustomerTier) String() string {
	switch t {
	case TierBronze:
		return "Bronze"
	case TierSilver:
		return "Silver"
	case TierGold:
		return "Gold"
	case TierPlatinum:
		return "Platinum"
	case TierDiamond:
		return "Diamond"
	default:
		return "Bronze"
	}
}

// CustomerTierIcon returns emoji icon for tier
func (t CustomerTier) Icon() string {
	switch t {
	case TierBronze:
		return "🥉"
	case TierSilver:
		return "🥈"
	case TierGold:
		return "🥇"
	case TierPlatinum:
		return "💎"
	case TierDiamond:
		return "💍"
	default:
		return "🥉"
	}
}

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
		Tier:      TierBronze,
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
	
	// Basic customer info
	Phone           string       `json:"phone" db:"phone"`
	FirstName       string       `json:"first_name" db:"first_name"`
	LastName        string       `json:"last_name" db:"last_name"`
	Email           string       `json:"email" db:"email"`
	DateOfBirth     *time.Time   `json:"date_of_birth" db:"date_of_birth"`
	Gender          *string      `json:"gender" db:"gender"`
	
	// SAAN-specific fields (ไม่ sync กับ Loyverse)
	CustomerCode    string       `json:"customer_code" db:"customer_code"` // SAAN00012345
	Tier            CustomerTier `json:"tier" db:"tier"`
	PointsBalance   int          `json:"points_balance" db:"points_balance"`
	TotalSpent      float64      `json:"total_spent" db:"total_spent"`
	TierAchievedDate *time.Time  `json:"tier_achieved_date" db:"tier_achieved_date"`
	
	// Loyverse integration fields
	LoyverseID          *string    `json:"loyverse_id" db:"loyverse_id"`
	LoyverseTotalVisits int        `json:"loyverse_total_visits" db:"loyverse_total_visits"`
	LoyverseTotalSpent  float64    `json:"loyverse_total_spent" db:"loyverse_total_spent"`
	LoyversePoints      int        `json:"loyverse_points" db:"loyverse_points"`
	FirstVisit          *time.Time `json:"first_visit" db:"first_visit"`
	LastVisit           *time.Time `json:"last_visit" db:"last_visit"`
	LastSyncAt          *time.Time `json:"last_sync_at" db:"last_sync_at"`
	
	// LINE integration
	LineUserID          *string    `json:"line_user_id" db:"line_user_id"`
	LineDisplayName     *string    `json:"line_display_name" db:"line_display_name"`
	DigitalCardIssuedAt *time.Time `json:"digital_card_issued_at" db:"digital_card_issued_at"`
	LastCardScan        *time.Time `json:"last_card_scan" db:"last_card_scan"`
	
	// Purchase analytics (denormalized for performance)
	OrderCount          int        `json:"order_count" db:"order_count"`
	LastOrderDate       *time.Time `json:"last_order_date" db:"last_order_date"`
	AverageOrderValue   float64    `json:"average_order_value" db:"average_order_value"`
	PurchaseFrequency   *float64   `json:"purchase_frequency" db:"purchase_frequency"` // orders per month
	
	// Delivery routing
	DeliveryRouteID     *uuid.UUID `json:"delivery_route_id" db:"delivery_route_id"`
	
	// System fields
	IsActive            bool       `json:"is_active" db:"is_active"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
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
	Customer  *Customer          `json:"customer"`
	Addresses []CustomerAddress  `json:"addresses"`
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

// GetTierInfo returns tier information with benefits
func (c *Customer) GetTierInfo() VIPTierInfo {
	return VIPTierInfo{
		Level:    int(c.Tier),
		Name:     c.Tier.String(),
		Icon:     c.Tier.Icon(),
		MinSpent: GetTierMinSpent(c.Tier),
	}
}

// CanUpgradeTier checks if customer can upgrade to next tier
func (c *Customer) CanUpgradeTier() (bool, CustomerTier) {
	nextTier := c.Tier + 1
	if nextTier > TierDiamond {
		return false, c.Tier
	}
	
	minSpent := GetTierMinSpent(nextTier)
	return c.TotalSpent >= minSpent, nextTier
}

// EarnPoints adds points to customer balance
func (c *Customer) EarnPoints(points int) {
	c.PointsBalance += points
	c.UpdatedAt = time.Now()
}

// RedeemPoints subtracts points from customer balance
func (c *Customer) RedeemPoints(points int) error {
	if c.PointsBalance < points {
		return ErrInsufficientPoints
	}
	c.PointsBalance -= points
	c.UpdatedAt = time.Now()
	return nil
}

// UpdateTotalSpent updates total spent and checks for tier upgrade
func (c *Customer) UpdateTotalSpent(amount float64) (bool, CustomerTier) {
	oldTier := c.Tier
	c.TotalSpent += amount
	
	// Check for tier upgrade
	newTier := CalculateTierFromSpending(c.TotalSpent)
	if newTier > oldTier {
		c.Tier = newTier
		now := time.Now()
		c.TierAchievedDate = &now
		c.UpdatedAt = now
		return true, newTier
	}
	
	c.UpdatedAt = time.Now()
	return false, oldTier
}

// IssueLINEDigitalCard issues a LINE digital card
func (c *Customer) IssueLINEDigitalCard() {
	now := time.Now()
	c.DigitalCardIssuedAt = &now
	c.UpdatedAt = now
}

// RecordCardScan records a card scan event
func (c *Customer) RecordCardScan() {
	now := time.Now()
	c.LastCardScan = &now
	c.UpdatedAt = now
}

// GenerateCustomerCode generates a unique customer code
func GenerateCustomerCode() string {
	// Simple implementation - can be improved with sequence
	return fmt.Sprintf("SAAN%06d", time.Now().Unix()%1000000)
}

// CalculateTierFromSpent calculates the tier based on total spent amount
func CalculateTierFromSpent(totalSpent float64) int {
	if totalSpent >= 250000 {
		return 5 // Diamond
	} else if totalSpent >= 100000 {
		return 4 // Platinum
	} else if totalSpent >= 50000 {
		return 3 // Gold
	} else if totalSpent >= 10000 {
		return 2 // Silver
	}
	return 1 // Bronze
}

// CustomerPointsStats represents customer points statistics
type CustomerPointsStats struct {
	CurrentBalance     int `json:"current_balance"`
	TotalEarned        int `json:"total_earned"`
	TotalRedeemed      int `json:"total_redeemed"`
	TotalTransactions  int `json:"total_transactions"`
}

// VIPTierInfo represents VIP tier information
type VIPTierInfo struct {
	Level    int     `json:"level"`
	Name     string  `json:"name"`
	Icon     string  `json:"icon"`
	MinSpent float64 `json:"min_spent"`
}

// GetTierMinSpent returns minimum spending for tier
func GetTierMinSpent(tier CustomerTier) float64 {
	switch tier {
	case TierBronze:
		return 0
	case TierSilver:
		return 10000
	case TierGold:
		return 50000
	case TierPlatinum:
		return 100000
	case TierDiamond:
		return 250000
	default:
		return 0
	}
}

// CalculateTierFromSpending calculates tier based on total spending
func CalculateTierFromSpending(totalSpent float64) CustomerTier {
	if totalSpent >= 250000 {
		return TierDiamond
	} else if totalSpent >= 100000 {
		return TierPlatinum
	} else if totalSpent >= 50000 {
		return TierGold
	} else if totalSpent >= 10000 {
		return TierSilver
	}
	return TierBronze
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

// NewCustomerAddress creates a new customer address
func NewCustomerAddress(
	customerID uuid.UUID,
	addressType AddressType,
	label, addressLine1, addressLine2, subDistrict, district, province, postalCode string,
	isDefault bool,
) (*CustomerAddress, error) {
	address := &CustomerAddress{
		ID:           uuid.New(),
		CustomerID:   customerID,
		Type:         string(addressType),
		Label:        label,
		AddressLine1: addressLine1,
		SubDistrict:  subDistrict,
		District:     district,
		Province:     province,
		PostalCode:   postalCode,
		IsDefault:    isDefault,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if addressLine2 != "" {
		address.AddressLine2 = &addressLine2
	}

	// Validate the address
	if err := address.Validate(); err != nil {
		return nil, err
	}

	return address, nil
}

// Validate validates customer address data
func (a *CustomerAddress) Validate() error {
	if a.CustomerID == uuid.Nil {
		return ErrInvalidCustomerID
	}
	if a.AddressLine1 == "" {
		return ErrInvalidAddressLine1
	}
	if a.SubDistrict == "" {
		return ErrInvalidSubDistrict
	}
	if a.District == "" {
		return ErrInvalidDistrict
	}
	if a.Province == "" {
		return ErrInvalidProvince
	}
	if a.PostalCode == "" {
		return ErrInvalidPostalCode
	}
	return nil
}

// Missing types for repository interfaces
type AddressSuggestion struct {
	ID          uuid.UUID `json:"id"`
	FullAddress string    `json:"full_address"`
	Province    string    `json:"province"`
	District    string    `json:"district"`
	Subdistrict string    `json:"subdistrict"`
	PostalCode  string    `json:"postal_code"`
	Relevance   float64   `json:"relevance"`
}

// Loyverse integration types
type LoyverseReceipt struct {
	ID            string    `json:"id"`
	ReceiptNumber string    `json:"receipt_number"`
	CustomerID    string    `json:"customer_id"`
	TotalMoney    float64   `json:"total_money"`
	PointsEarned  float64   `json:"points_earned"`
	CreatedAt     time.Time `json:"created_at"`
}

type LoyverseReceiptWebhook struct {
	ReceiptNumber       string    `json:"receipt_number"`
	CustomerID          string    `json:"customer_id"`
	TotalMoney          float64   `json:"total_money"`
	PointsEarned        float64   `json:"points_earned"`
	PointsBalance       float64   `json:"points_balance"`
	CustomerTotalVisits int       `json:"customer_total_visits"`
	CustomerTotalSpent  float64   `json:"customer_total_spent"`
	ReceiptDate         time.Time `json:"created_at"`
}

// VIP Tier Benefits System
type VIPTierBenefits struct {
	Tier                  CustomerTier `json:"tier" db:"tier"`
	TierName              string       `json:"tier_name" db:"tier_name"`
	TierIcon              string       `json:"tier_icon" db:"tier_icon"`
	MinSpent              float64      `json:"min_spent" db:"min_spent"`
	DiscountPercentage    float64      `json:"discount_percentage" db:"discount_percentage"`
	PointsMultiplier      float64      `json:"points_multiplier" db:"points_multiplier"`
	FreeShippingThreshold float64      `json:"free_shipping_threshold" db:"free_shipping_threshold"`
	SpecialOffers         []string     `json:"special_offers" db:"special_offers"`
	PrioritySupport       bool         `json:"priority_support" db:"priority_support"`
	EarlyAccess           bool         `json:"early_access" db:"early_access"`
	BirthdayBonus         int          `json:"birthday_bonus" db:"birthday_bonus"`
	ReferralBonus         int          `json:"referral_bonus" db:"referral_bonus"`
	CreatedAt             time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at" db:"updated_at"`
}

// Customer Points Transaction System
type CustomerPointsTransaction struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	CustomerID    uuid.UUID              `json:"customer_id" db:"customer_id"`
	TransactionID uuid.UUID              `json:"transaction_id" db:"transaction_id"`
	Type          PointsTransactionType  `json:"type" db:"type"`
	Points        int                    `json:"points" db:"points"`
	Balance       int                    `json:"balance" db:"balance"`
	Source        string                 `json:"source" db:"source"`
	Description   string                 `json:"description" db:"description"`
	ReferenceID   *uuid.UUID             `json:"reference_id" db:"reference_id"`
	ReferenceType *string                `json:"reference_type" db:"reference_type"`
	ExpiryDate    *time.Time             `json:"expiry_date" db:"expiry_date"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
}

// Points transaction types
type PointsTransactionType string

const (
	PointsEarned   PointsTransactionType = "earned"
	PointsRedeemed PointsTransactionType = "redeemed"
	PointsExpired  PointsTransactionType = "expired"
	PointsAdjusted PointsTransactionType = "adjusted"
)

// CustomerAnalytics represents customer analytics data
type CustomerAnalytics struct {
	ID                   uuid.UUID              `json:"id" db:"id"`
	CustomerID           uuid.UUID              `json:"customer_id" db:"customer_id"`
	AnalyticsDate        time.Time              `json:"analytics_date" db:"analytics_date"`
	TotalOrders          int                    `json:"total_orders" db:"total_orders"`
	TotalSpent           float64                `json:"total_spent" db:"total_spent"`
	AverageOrderValue    float64                `json:"average_order_value" db:"average_order_value"`
	PurchaseFrequency    float64                `json:"purchase_frequency" db:"purchase_frequency"`
	DaysSinceLastOrder   *int                   `json:"days_since_last_order" db:"days_since_last_order"`
	PreferredCategories  []string               `json:"preferred_categories" db:"preferred_categories"`
	PreferredBrands      []string               `json:"preferred_brands" db:"preferred_brands"`
	SeasonalTrends       map[string]interface{} `json:"seasonal_trends" db:"seasonal_trends"`
	CustomerSegment      string                 `json:"customer_segment" db:"customer_segment"`
	LifetimeValue        float64                `json:"lifetime_value" db:"lifetime_value"`
	ChurnRisk           float64                `json:"churn_risk" db:"churn_risk"`
	NextOrderPrediction  *time.Time             `json:"next_order_prediction" db:"next_order_prediction"`
	CreatedAt           time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time              `json:"updated_at" db:"updated_at"`
}

// UpsellSuggestion represents product upsell suggestions for customers
type UpsellSuggestion struct {
	ID            uuid.UUID `json:"id" db:"id"`
	CustomerID    uuid.UUID `json:"customer_id" db:"customer_id"`
	ProductID     uuid.UUID `json:"product_id" db:"product_id"`
	ProductName   string    `json:"product_name" db:"product_name"`
	Reason        string    `json:"reason" db:"reason"`
	Confidence    float64   `json:"confidence" db:"confidence"`
	PredictedCLV  float64   `json:"predicted_clv" db:"predicted_clv"`
	ValidUntil    time.Time `json:"valid_until" db:"valid_until"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}
