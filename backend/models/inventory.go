package models

import (
	"gorm.io/gorm"
)

type Inventory struct {
	gorm.Model
	RouteID    uint   `json:"route_id"`
	Date       string `json:"date"`
	TotalSpots int    `json:"total_spots"`
	LeftSpots  int    `json:"left_spots"`
	Status     string `json:"status"`
	Route      Route  `json:"route" gorm:"foreignKey:RouteID"`
}

type InventoryAdjustLog struct {
	gorm.Model
	InventoryID  uint      `json:"inventory_id"`
	RouteID      uint      `json:"route_id"`
	AdjustType   string    `json:"adjust_type"`
	OldSpots     int       `json:"old_spots"`
	NewSpots     int       `json:"new_spots"`
	AdjustAmount int       `json:"adjust_amount"`
	Reason       string    `json:"reason"`
	OrderID      *uint     `json:"order_id"`
	RefundID     *uint     `json:"refund_id"`
	Inventory    Inventory `json:"inventory" gorm:"foreignKey:InventoryID"`
}
