package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `json:"username" gorm:"uniqueIndex"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
}

type Trip struct {
	gorm.Model
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Destination string          `json:"destination"`
	StartDate   time.Time       `json:"start_date"`
	EndDate     time.Time       `json:"end_date"`
	Price       float64         `json:"price"`
	TotalSpots  int             `json:"total_spots"`
	LeftSpots   int             `json:"left_spots"`
	Status      string          `json:"status"`
	Itineraries []TripItinerary `json:"itineraries,omitempty" gorm:"foreignKey:TripID"`
}

type Order struct {
	gorm.Model
	OrderNo     string    `json:"order_no" gorm:"uniqueIndex"`
	UserID      *uint     `json:"user_id"`
	TripID      uint      `json:"trip_id"`
	TripName    string    `json:"trip_name"`
	TripPrice   float64   `json:"trip_price"`
	Travelers   int       `json:"travelers"`
	TotalAmount float64   `json:"total_amount"`
	ContactName string    `json:"contact_name"`
	ContactPhone string   `json:"contact_phone"`
	Status      string    `json:"status"`
	PayTime     *time.Time `json:"pay_time"`
	Trip        Trip      `json:"trip" gorm:"foreignKey:TripID"`
	User        *User     `json:"user" gorm:"foreignKey:UserID"`
}

type RefundRequest struct {
	gorm.Model
	RefundNo     string    `json:"refund_no" gorm:"uniqueIndex"`
	OrderID      uint      `json:"order_id"`
	OrderNo      string    `json:"order_no"`
	UserID       *uint     `json:"user_id"`
	Reason       string    `json:"reason"`
	Description  string    `json:"description"`
	RefundAmount float64   `json:"refund_amount"`
	Status       string    `json:"status"`
	ReviewerID   *uint     `json:"reviewer_id"`
	ReviewRemark string    `json:"review_remark"`
	ReviewTime   *time.Time `json:"review_time"`
	Order        Order     `json:"order" gorm:"foreignKey:OrderID"`
	User         *User     `json:"user" gorm:"foreignKey:UserID"`
	ReviewLogs   []RefundReviewLog `json:"review_logs,omitempty" gorm:"foreignKey:RefundID"`
}

type RefundReviewLog struct {
	gorm.Model
	RefundID   uint   `json:"refund_id"`
	Action     string `json:"action"`
	FromStatus string `json:"from_status"`
	ToStatus   string `json:"to_status"`
	Remark     string `json:"remark"`
	OperatorID *uint  `json:"operator_id"`
	Refund     RefundRequest `json:"-" gorm:"foreignKey:RefundID"`
	Operator   *User  `json:"operator" gorm:"foreignKey:OperatorID"`
}

type TripItinerary struct {
	gorm.Model
	TripID     uint   `json:"trip_id"`
	DayNumber  int    `json:"day_number"`
	Title      string `json:"title"`
	Breakfast  string `json:"breakfast"`
	Lunch      string `json:"lunch"`
	Dinner     string `json:"dinner"`
	Accommodation string `json:"accommodation"`
	Transportation string `json:"transportation"`
	Activities string `json:"activities"`
	Notes      string `json:"notes"`
	Trip       Trip   `json:"trip" gorm:"foreignKey:TripID"`
}

type SpotAdjustmentLog struct {
	gorm.Model
	TripID        uint   `json:"trip_id"`
	AdjustType    string `json:"adjust_type"`
	OldSpots      int    `json:"old_spots"`
	NewSpots      int    `json:"new_spots"`
	AdjustAmount  int    `json:"adjust_amount"`
	Reason        string `json:"reason"`
	OperatorID    *uint  `json:"operator_id"`
	OrderID       *uint  `json:"order_id"`
	RefundID      *uint  `json:"refund_id"`
	Trip          Trip   `json:"trip" gorm:"foreignKey:TripID"`
}
