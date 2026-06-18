package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	OrderNo      string    `json:"order_no" gorm:"uniqueIndex"`
	UserID       *uint     `json:"user_id"`
	RouteID      uint      `json:"route_id"`
	TourGroupID  uint      `json:"tour_group_id"`
	RouteName    string    `json:"route_name"`
	RoutePrice   float64   `json:"route_price"`
	TravelerCount int      `json:"traveler_count"`
	TotalAmount  float64   `json:"total_amount"`
	ContactName  string    `json:"contact_name"`
	ContactPhone string    `json:"contact_phone"`
	ContactEmail string    `json:"contact_email"`
	Status       string    `json:"status"`
	PayTime      *time.Time `json:"pay_time"`
	PayMethod    string    `json:"pay_method"`
	Remark       string    `json:"remark"`
	Route        Route     `json:"route" gorm:"foreignKey:RouteID"`
	TourGroup    TourGroup `json:"tour_group" gorm:"foreignKey:TourGroupID"`
	User         *User     `json:"user" gorm:"foreignKey:UserID"`
	Travelers    []Traveler `json:"travelers,omitempty" gorm:"foreignKey:OrderID"`
}

type Traveler struct {
	gorm.Model
	OrderID      uint   `json:"order_id"`
	Name         string `json:"name"`
	IDCardNo       string `json:"id_card_no"`
	Phone        string `json:"phone"`
	Gender       string `json:"gender"`
	BirthDate    string `json:"birth_date"`
	Nationality  string `json:"nationality"`
	Remark       string `json:"remark"`
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
	Operator   *User  `json:"operator" gorm:"foreignKey:OperatorID"`
}
