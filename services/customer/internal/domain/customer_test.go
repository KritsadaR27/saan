package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCustomer(t *testing.T) {
	tests := []struct {
		name          string
		email         string
		phone         string
		firstName     string
		lastName      string
		expectError   bool
		errorContains string
	}{
		{
			name:      "valid customer",
			email:     "test@example.com",
			phone:     "0812345678",
			firstName: "John",
			lastName:  "Doe",
		},
		{
			name:          "empty email",
			email:         "",
			phone:         "0812345678",
			firstName:     "John",
			lastName:      "Doe",
			expectError:   true,
			errorContains: "invalid email format",
		},
		{
			name:          "invalid email",
			email:         "invalid-email",
			phone:         "0812345678",
			firstName:     "John",
			lastName:      "Doe",
			expectError:   true,
			errorContains: "invalid email format",
		},
		{
			name:          "empty phone",
			email:         "test@example.com",
			phone:         "",
			firstName:     "John",
			lastName:      "Doe",
			expectError:   true,
			errorContains: "invalid phone format",
		},
		{
			name:          "invalid phone format",
			email:         "test@example.com",
			phone:         "123",
			firstName:     "John",
			lastName:      "Doe",
			expectError:   true,
			errorContains: "invalid phone format",
		},
		{
			name:          "empty first name",
			email:         "test@example.com",
			phone:         "0812345678",
			firstName:     "",
			lastName:      "Doe",
			expectError:   true,
			errorContains: "first name is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			customer, err := NewCustomer(tt.email, tt.phone, tt.firstName, tt.lastName)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, customer)
			} else {
				require.NoError(t, err)
				require.NotNil(t, customer)
				assert.NotEmpty(t, customer.ID)
				assert.Equal(t, tt.email, customer.Email)
				assert.Equal(t, tt.phone, customer.Phone)
				assert.Equal(t, tt.firstName, customer.FirstName)
				assert.Equal(t, tt.lastName, customer.LastName)
				assert.Equal(t, CustomerTierBronze, customer.Tier)
				assert.True(t, customer.IsActive)
				assert.NotZero(t, customer.CreatedAt)
				assert.NotZero(t, customer.UpdatedAt)
			}
		})
	}
}

func TestCustomer_UpdateProfile(t *testing.T) {
	customer, err := NewCustomer("test@example.com", "0812345678", "John", "Doe")
	require.NoError(t, err)

	originalUpdatedAt := customer.UpdatedAt
	time.Sleep(time.Millisecond) // Ensure different timestamp

	err = customer.UpdateProfile("jane@example.com", "0887654321", "Jane", "Smith")
	require.NoError(t, err)

	assert.Equal(t, "jane@example.com", customer.Email)
	assert.Equal(t, "0887654321", customer.Phone)
	assert.Equal(t, "Jane", customer.FirstName)
	assert.Equal(t, "Smith", customer.LastName)
	assert.True(t, customer.UpdatedAt.After(originalUpdatedAt))
}

func TestCustomer_UpdateTier(t *testing.T) {
	customer, err := NewCustomer("test@example.com", "0812345678", "John", "Doe")
	require.NoError(t, err)

	originalUpdatedAt := customer.UpdatedAt
	time.Sleep(time.Millisecond)

	customer.UpdateTier(CustomerTierGold)

	assert.Equal(t, CustomerTierGold, customer.Tier)
	assert.True(t, customer.UpdatedAt.After(originalUpdatedAt))
}

func TestCustomer_SetLoyverseID(t *testing.T) {
	customer, err := NewCustomer("test@example.com", "0812345678", "John", "Doe")
	require.NoError(t, err)

	loyverseID := "loyverse_123"
	originalUpdatedAt := customer.UpdatedAt
	time.Sleep(time.Millisecond)

	customer.SetLoyverseID(loyverseID)

	assert.Equal(t, &loyverseID, customer.LoyverseID)
	assert.True(t, customer.UpdatedAt.After(originalUpdatedAt))
}

func TestCustomer_Deactivate(t *testing.T) {
	customer, err := NewCustomer("test@example.com", "0812345678", "John", "Doe")
	require.NoError(t, err)
	assert.True(t, customer.IsActive)

	originalUpdatedAt := customer.UpdatedAt
	time.Sleep(time.Millisecond)

	customer.Deactivate()

	assert.False(t, customer.IsActive)
	assert.True(t, customer.UpdatedAt.After(originalUpdatedAt))
}

func TestCustomer_Activate(t *testing.T) {
	customer, err := NewCustomer("test@example.com", "0812345678", "John", "Doe")
	require.NoError(t, err)

	customer.Deactivate()
	assert.False(t, customer.IsActive)

	originalUpdatedAt := customer.UpdatedAt
	time.Sleep(time.Millisecond)

	customer.Activate()

	assert.True(t, customer.IsActive)
	assert.True(t, customer.UpdatedAt.After(originalUpdatedAt))
}

func TestNewCustomerAddress(t *testing.T) {
	customerID := uuid.New()

	tests := []struct {
		name           string
		addressType    AddressType
		label          string
		addressLine1   string
		addressLine2   string
		subDistrict    string
		district       string
		province       string
		postalCode     string
		isDefault      bool
		expectError    bool
		errorContains  string
	}{
		{
			name:         "valid home address",
			addressType:  AddressTypeHome,
			label:        "My Home",
			addressLine1: "123 Main St",
			addressLine2: "Apt 4B",
			subDistrict:  "Khlong Toei",
			district:     "Khlong Toei",
			province:     "Bangkok",
			postalCode:   "10110",
			isDefault:    true,
		},
		{
			name:          "empty address line 1",
			addressType:   AddressTypeHome,
			label:         "My Home",
			addressLine1:  "",
			subDistrict:   "Khlong Toei",
			district:      "Khlong Toei",
			province:      "Bangkok",
			postalCode:    "10110",
			expectError:   true,
			errorContains: "address line 1 is required",
		},
		{
			name:          "empty sub district",
			addressType:   AddressTypeHome,
			label:         "My Home",
			addressLine1:  "123 Main St",
			subDistrict:   "",
			district:      "Khlong Toei",
			province:      "Bangkok",
			postalCode:    "10110",
			expectError:   true,
			errorContains: "sub district is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			address, err := NewCustomerAddress(
				customerID,
				tt.addressType,
				tt.label,
				tt.addressLine1,
				tt.addressLine2,
				tt.subDistrict,
				tt.district,
				tt.province,
				tt.postalCode,
				tt.isDefault,
			)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, address)
			} else {
				require.NoError(t, err)
				require.NotNil(t, address)
				assert.NotEmpty(t, address.ID)
				assert.Equal(t, customerID, address.CustomerID)
				assert.Equal(t, string(tt.addressType), address.Type)
				assert.Equal(t, tt.label, address.Label)
				assert.Equal(t, tt.addressLine1, address.AddressLine1)
				if tt.addressLine2 != "" {
					require.NotNil(t, address.AddressLine2)
					assert.Equal(t, tt.addressLine2, *address.AddressLine2)
				} else {
					assert.Nil(t, address.AddressLine2)
				}
				assert.Equal(t, tt.subDistrict, address.SubDistrict)
				assert.Equal(t, tt.district, address.District)
				assert.Equal(t, tt.province, address.Province)
				assert.Equal(t, tt.postalCode, address.PostalCode)
				assert.Equal(t, tt.isDefault, address.IsDefault)
				assert.NotZero(t, address.CreatedAt)
				assert.NotZero(t, address.UpdatedAt)
			}
		})
	}
}

func TestCustomerAddress_Update(t *testing.T) {
	customerID := uuid.New()
	address, err := NewCustomerAddress(
		customerID,
		AddressTypeHome,
		"My Home",
		"123 Main St",
		"",
		"Khlong Toei",
		"Khlong Toei",
		"Bangkok",
		"10110",
		true,
	)
	require.NoError(t, err)

	originalUpdatedAt := address.UpdatedAt
	time.Sleep(time.Millisecond)

	err = address.Update(
		AddressTypeWork,
		"My Office",
		"456 Business Ave",
		"Floor 10",
		"Silom",
		"Bang Rak",
		"Bangkok",
		"10500",
		false,
	)
	require.NoError(t, err)

	assert.Equal(t, string(AddressTypeWork), address.Type)
	assert.Equal(t, "My Office", address.Label)
	assert.Equal(t, "456 Business Ave", address.AddressLine1)
	require.NotNil(t, address.AddressLine2)
	assert.Equal(t, "Floor 10", *address.AddressLine2)
	assert.Equal(t, "Silom", address.SubDistrict)
	assert.Equal(t, "Bang Rak", address.District)
	assert.Equal(t, "Bangkok", address.Province)
	assert.Equal(t, "10500", address.PostalCode)
	assert.False(t, address.IsDefault)
	assert.True(t, address.UpdatedAt.After(originalUpdatedAt))
}

func TestCustomerAddress_SetAsDefault(t *testing.T) {
	customerID := uuid.New()
	address, err := NewCustomerAddress(
		customerID,
		AddressTypeHome,
		"My Home",
		"123 Main St",
		"",
		"Khlong Toei",
		"Khlong Toei",
		"Bangkok",
		"10110",
		false,
	)
	require.NoError(t, err)
	assert.False(t, address.IsDefault)

	originalUpdatedAt := address.UpdatedAt
	time.Sleep(time.Millisecond)

	address.SetAsDefault()

	assert.True(t, address.IsDefault)
	assert.True(t, address.UpdatedAt.After(originalUpdatedAt))
}

func TestCustomerAddress_UnsetAsDefault(t *testing.T) {
	customerID := uuid.New()
	address, err := NewCustomerAddress(
		customerID,
		AddressTypeHome,
		"My Home",
		"123 Main St",
		"",
		"Khlong Toei",
		"Khlong Toei",
		"Bangkok",
		"10110",
		true,
	)
	require.NoError(t, err)
	assert.True(t, address.IsDefault)

	originalUpdatedAt := address.UpdatedAt
	time.Sleep(time.Millisecond)

	address.UnsetAsDefault()

	assert.False(t, address.IsDefault)
	assert.True(t, address.UpdatedAt.After(originalUpdatedAt))
}
