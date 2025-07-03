package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DeliveryRoute represents delivery route mapping
type DeliveryRoute struct {
	ID              uint   `gorm:"primaryKey;autoIncrement"`
	SubdistrictTh   string `gorm:"column:subdistrict_th;size:100"`
	DistrictTh      string `gorm:"column:district_th;size:100"`
	ProvinceTh      string `gorm:"column:province_th;size:100"`
	Route           string `gorm:"column:route;size:10"`
	DeliveryDays    string `gorm:"column:delivery_days;size:50"` // "Monday,Thursday" etc.
}

func (DeliveryRoute) TableName() string {
	return "delivery_routes"
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL not set")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Create delivery_routes table
	if err := db.AutoMigrate(&DeliveryRoute{}); err != nil {
		log.Fatalf("failed to migrate delivery_routes table: %v", err)
	}

	// Clear existing data
	db.Exec("TRUNCATE TABLE delivery_routes RESTART IDENTITY")

	// Load CSV data
	file, err := os.Open("delivery_route.csv")
	if err != nil {
		log.Fatalf("failed to open delivery_route.csv: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("failed to read CSV: %v", err)
	}

	// Define delivery days mapping
	deliveryDaysMap := map[string]string{
		"A": "Tuesday,Friday",
		"B": "Tuesday,Friday", 
		"C": "Wednesday,Saturday",
		"D": "Monday,Thursday",
		"E": "Monday,Thursday",
		"F": "Tuesday,Friday",
		"G": "Wednesday,Saturday",
		"H": "Tuesday,Friday",
		"J": "Wednesday,Saturday",
	}

	var routes []DeliveryRoute

	// Skip header row
	for i := 1; i < len(records); i++ {
		record := records[i]
		if len(record) != 4 {
			continue
		}

		subdistrict := strings.TrimSpace(record[0])
		district := strings.TrimSpace(record[1])
		province := strings.TrimSpace(record[2])
		routeCode := strings.TrimSpace(record[3])

		deliveryDays, exists := deliveryDaysMap[routeCode]
		if !exists {
			deliveryDays = "Tuesday,Friday" // default
		}

		route := DeliveryRoute{
			SubdistrictTh: subdistrict,
			DistrictTh:    district,
			ProvinceTh:    province,
			Route:         routeCode,
			DeliveryDays:  deliveryDays,
		}

		routes = append(routes, route)
	}

	// Batch insert
	batchSize := 100
	for i := 0; i < len(routes); i += batchSize {
		end := i + batchSize
		if end > len(routes) {
			end = len(routes)
		}

		if err := db.Create(routes[i:end]).Error; err != nil {
			log.Fatalf("failed to insert delivery routes batch: %v", err)
		}
	}

	fmt.Printf("✅ Successfully imported delivery route data into PostgreSQL\n")
	fmt.Printf("   - Total routes: %d\n", len(routes))

	// Show summary by route
	var summary []struct {
		Route string
		Count int64
		DeliveryDays string
	}

	db.Raw(`
		SELECT route, COUNT(*) as count, delivery_days 
		FROM delivery_routes 
		GROUP BY route, delivery_days 
		ORDER BY route
	`).Scan(&summary)

	fmt.Println("\n📊 Route Summary:")
	for _, s := range summary {
		fmt.Printf("   Route %s: %d areas (%s)\n", s.Route, s.Count, s.DeliveryDays)
	}
}
