package importaddresses

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Province struct {
	ProvinceCode int    `json:"provinceCode" gorm:"primaryKey"`
	NameTh       string `json:"provinceNameTh"`
	NameEn       string `json:"provinceNameEn"`
}

type District struct {
	DistrictCode int    `json:"districtCode" gorm:"primaryKey"`
	ProvinceCode int    `json:"provinceCode"`
	NameTh       string `json:"districtNameTh"`
	NameEn       string `json:"districtNameEn"`
	PostalCode   string `json:"postalCode"`
}

type Subdistrict struct {
	SubdistrictCode int    `json:"subdistrictCode" gorm:"primaryKey"`
	DistrictCode    int    `json:"districtCode"`
	ProvinceCode    int    `json:"provinceCode"`
	NameTh          string `json:"subdistrictNameTh"`
	NameEn          string `json:"subdistrictNameEn"`
	PostalCode      string `json:"postalCode"`
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

	db.AutoMigrate(&Province{}, &District{}, &Subdistrict{})
	db.Exec("TRUNCATE TABLE subdistricts, districts, provinces RESTART IDENTITY CASCADE")

	base := filepath.Dir(os.Args[0])

	provinces, err := loadJSON[Province](filepath.Join(base, "provinces.json"))
	if err != nil {
		log.Fatal("failed to load provinces: ", err)
	}
	districts, err := loadJSON[District](filepath.Join(base, "districts.json"))
	if err != nil {
		log.Fatal("failed to load districts: ", err)
	}
	subdistricts, err := loadJSON[Subdistrict](filepath.Join(base, "subdistricts.json"))
	if err != nil {
		log.Fatal("failed to load subdistricts: ", err)
	}

	db.Create(&provinces)
	db.Create(&districts)
	db.Create(&subdistricts)

	fmt.Println("✅ Successfully imported Thai address data into PostgreSQL")
}
