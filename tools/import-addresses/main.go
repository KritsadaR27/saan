package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Province struct {
	ProvinceCode int    `json:"provinceCode" gorm:"primaryKey;column:province_code"`
	NameTh       string `json:"provinceNameTh" gorm:"column:name_th"`
	NameEn       string `json:"provinceNameEn" gorm:"column:name_en"`
}

func (Province) TableName() string {
	return "provinces"
}

type District struct {
	DistrictCode int    `json:"districtCode" gorm:"primaryKey;column:district_code"`
	ProvinceCode int    `json:"provinceCode" gorm:"column:province_code"`
	NameTh       string `json:"districtNameTh" gorm:"column:name_th"`
	NameEn       string `json:"districtNameEn" gorm:"column:name_en"`
	PostalCode   int    `json:"postalCode" gorm:"column:postal_code"`
}

func (District) TableName() string {
	return "districts"
}

type Subdistrict struct {
	SubdistrictCode int    `json:"subdistrictCode" gorm:"primaryKey;column:subdistrict_code"`
	DistrictCode    int    `json:"districtCode" gorm:"column:district_code"`
	ProvinceCode    int    `json:"provinceCode" gorm:"column:province_code"`
	NameTh          string `json:"subdistrictNameTh" gorm:"column:name_th"`
	NameEn          string `json:"subdistrictNameEn" gorm:"column:name_en"`
	PostalCode      int    `json:"postalCode" gorm:"column:postal_code"`
}

func (Subdistrict) TableName() string {
	return "subdistricts"
}

func loadJSON[T any](filename string) ([]T, error) {
	var data []T
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	return data, err
}

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Println("No .env file found, using system environment")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=password dbname=saan_db port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Create tables if not exist
	err = db.AutoMigrate(&Province{}, &District{}, &Subdistrict{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// Clear existing data
	db.Exec("TRUNCATE TABLE subdistricts, districts, provinces RESTART IDENTITY CASCADE")

	// Load data files
	provinces, err := loadJSON[Province]("provinces.json")
	if err != nil {
		log.Fatal("failed to load provinces: ", err)
	}
	
	districts, err := loadJSON[District]("districts.json")
	if err != nil {
		log.Fatal("failed to load districts: ", err)
	}
	
	subdistricts, err := loadJSON[Subdistrict]("subdistricts.json")
	if err != nil {
		log.Fatal("failed to load subdistricts: ", err)
	}

	// Insert data
	if err := db.Create(&provinces).Error; err != nil {
		log.Fatal("failed to insert provinces: ", err)
	}
	
	if err := db.Create(&districts).Error; err != nil {
		log.Fatal("failed to insert districts: ", err)
	}
	
	if err := db.Create(&subdistricts).Error; err != nil {
		log.Fatal("failed to insert subdistricts: ", err)
	}

	fmt.Printf("✅ Successfully imported Thai address data into PostgreSQL\n")
	fmt.Printf("   - Provinces: %d\n", len(provinces))
	fmt.Printf("   - Districts: %d\n", len(districts))
	fmt.Printf("   - Subdistricts: %d\n", len(subdistricts))
}
