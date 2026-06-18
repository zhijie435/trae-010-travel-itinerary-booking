package models

import (
	"time"

	"gorm.io/gorm"
)

type Route struct {
	gorm.Model
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Destination string      `json:"destination"`
	StartDate   time.Time   `json:"start_date"`
	EndDate     time.Time   `json:"end_date"`
	Days        int         `json:"days"`
	Nights      int         `json:"nights"`
	Price       float64     `json:"price"`
	Status      string      `json:"status"`
	Itineraries []Itinerary `json:"itineraries,omitempty" gorm:"foreignKey:RouteID"`
}

type Itinerary struct {
	gorm.Model
	RouteID        uint   `json:"route_id"`
	DayNumber      int    `json:"day_number"`
	Title          string `json:"title"`
	Breakfast      string `json:"breakfast"`
	Lunch          string `json:"lunch"`
	Dinner         string `json:"dinner"`
	Accommodation  string `json:"accommodation"`
	Transportation string `json:"transportation"`
	Activities     string `json:"activities"`
	Notes          string `json:"notes"`
}
