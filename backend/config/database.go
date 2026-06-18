package config

import (
	"log"
	"travel-refund/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("travel.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.Trip{},
		&models.Order{},
		&models.RefundRequest{},
		&models.RefundReviewLog{},
		&models.TripItinerary{},
		&models.SpotAdjustmentLog{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database initialized successfully")
}
