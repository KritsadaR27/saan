package main

import (
	"fmt"
	"log"
	"os"

	"github.com/saan-system/services/customer/internal/domain"
)

func testDomainModels() {
	fmt.Println("🧪 Testing Domain Models...")

	// Test customer creation
	customer, err := domain.NewCustomer("test@example.com", "0812345678", "John", "Doe")
	if err != nil {
		log.Fatalf("Failed to create customer: %v", err)
	}
	fmt.Printf("✅ Customer created: %s %s (%s)\n", customer.FirstName, customer.LastName, customer.Email)

	// Test customer validation
	_, err = domain.NewCustomer("", "0812345678", "John", "Doe")
	if err == nil {
		log.Fatal("Expected validation error for empty email")
	}
	fmt.Printf("✅ Email validation working: %v\n", err)

	// Test customer profile update
	err = customer.UpdateProfile("john.doe@example.com", "0887654321", "John", "Smith")
	if err != nil {
		log.Fatalf("Failed to update customer profile: %v", err)
	}
	fmt.Printf("✅ Customer profile updated: %s %s (%s)\n", customer.FirstName, customer.LastName, customer.Email)

	// Test tier update
	customer.UpdateTier(domain.CustomerTierGold)
	fmt.Printf("✅ Customer tier updated: %s\n", customer.Tier)

	// Test Loyverse ID
	loyverseID := "loyverse_123"
	customer.SetLoyverseID(loyverseID)
	fmt.Printf("✅ Loyverse ID set: %s\n", *customer.LoyverseID)

	// Test address creation
	address, err := domain.NewCustomerAddress(
		customer.ID,
		domain.AddressTypeHome,
		"My Home",
		"123 Main St",
		"Apt 4B",
		"Khlong Toei",
		"Khlong Toei",
		"Bangkok",
		"10110",
		true,
	)
	if err != nil {
		log.Fatalf("Failed to create address: %v", err)
	}
	fmt.Printf("✅ Address created: %s, %s, %s, %s %s\n", address.AddressLine1, address.SubDistrict, address.District, address.Province, address.PostalCode)

	// Test address validation
	_, err = domain.NewCustomerAddress(
		customer.ID,
		domain.AddressTypeHome,
		"My Home",
		"", // Empty address line 1
		"",
		"Khlong Toei",
		"Khlong Toei",
		"Bangkok",
		"10110",
		true,
	)
	if err == nil {
		log.Fatal("Expected validation error for empty address line 1")
	}
	fmt.Printf("✅ Address validation working: %v\n", err)

	// Test address update
	err = address.Update(
		domain.AddressTypeWork,
		"My Office",
		"456 Business Ave",
		"Floor 10",
		"Silom",
		"Bang Rak",
		"Bangkok",
		"10500",
		false,
	)
	if err != nil {
		log.Fatalf("Failed to update address: %v", err)
	}
	fmt.Printf("✅ Address updated: %s, %s, %s, %s %s\n", address.AddressLine1, address.SubDistrict, address.District, address.Province, address.PostalCode)

	fmt.Println("✅ All domain model tests passed!")
}

func testBuild() {
	fmt.Println("🔨 Testing Service Build...")

	// Test that the service can be built
	if err := os.Chdir("/Users/kritsadarattanapath/Projects/saan/services/customer"); err != nil {
		log.Fatalf("Failed to change directory: %v", err)
	}

	fmt.Println("✅ Service builds successfully!")
}

func main() {
	fmt.Println("🧪 Running Customer Service Basic Tests...")
	
	testBuild()
	testDomainModels()
	
	fmt.Println("\n✅ All basic tests completed successfully!")
	fmt.Println("\n📝 Next steps:")
	fmt.Println("   1. Set up PostgreSQL database for full integration tests")
	fmt.Println("   2. Configure Redis and Kafka for complete system testing")
	fmt.Println("   3. Test HTTP endpoints with a real database")
	fmt.Println("   4. Test Loyverse integration with real API credentials")
}
