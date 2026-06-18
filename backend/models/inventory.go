package models

import (
	"time"

	"gorm.io/gorm"
)

type TourGroup struct {
	gorm.Model
	RouteID      uint      `json:"route_id"`
	GroupNo      string    `json:"group_no" gorm:"uniqueIndex"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	TotalSpots   int       `json:"total_spots"`
	LeftSpots    int       `json:"left_spots"`
	Status       string    `json:"status"`
	GuideName    string    `json:"guide_name"`
	GuidePhone   string    `json:"guide_phone"`
	MeetingPoint string    `json:"meeting_point"`
	Remark       string    `json:"remark"`
	Route        Route     `json:"route" gorm:"foreignKey:RouteID"`
	Orders       []Order   `json:"orders,omitempty" gorm:"foreignKey:TourGroupID"`
}

type TourGroupAdjustLog struct {
	gorm.Model
	TourGroupID  uint      `json:"tour_group_id"`
	RouteID      uint      `json:"route_id"`
	AdjustType   string    `json:"adjust_type"`
	OldSpots     int       `json:"old_spots"`
	NewSpots     int       `json:"new_spots"`
	AdjustAmount int       `json:"adjust_amount"`
	Reason       string    `json:"reason"`
	OperatorID   *uint     `json:"operator_id"`
	OrderID      *uint     `json:"order_id"`
	RefundID     *uint     `json:"refund_id"`
	AdjustTime   time.Time `json:"adjust_time"`
	TourGroup    TourGroup `json:"tour_group" gorm:"foreignKey:TourGroupID"`
	Operator     *User     `json:"operator" gorm:"foreignKey:OperatorID"`
}
