package controllers

import (
	"net/http"
	"travel-refund/config"
	"travel-refund/models"

	"github.com/gin-gonic/gin"
)

type RefundRequestInput struct {
	OrderID     uint    `json:"order_id" binding:"required"`
	Reason      string  `json:"reason" binding:"required"`
	Description string  `json:"description"`
	RefundAmount float64 `json:"refund_amount"`
}

type ReviewInput struct {
	Status       string `json:"status" binding:"required"`
	ReviewRemark string `json:"review_remark"`
}

type TripInput struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Destination string  `json:"destination" binding:"required"`
	StartDate   string  `json:"start_date" binding:"required"`
	EndDate     string  `json:"end_date" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	TotalSpots  int     `json:"total_spots" binding:"required"`
	Status      string  `json:"status"`
}

type SpotAdjustInput struct {
	AdjustAmount int    `json:"adjust_amount" binding:"required"`
	Reason       string `json:"reason" binding:"required"`
}

type ItineraryInput struct {
	DayNumber       int    `json:"day_number" binding:"required"`
	Title           string `json:"title" binding:"required"`
	Breakfast       string `json:"breakfast"`
	Lunch           string `json:"lunch"`
	Dinner          string `json:"dinner"`
	Accommodation   string `json:"accommodation"`
	Transportation  string `json:"transportation"`
	Activities      string `json:"activities"`
	Notes           string `json:"notes"`
}

type BatchReviewInput struct {
	IDs          []uint `json:"ids" binding:"required,min=1"`
	Status       string `json:"status" binding:"required"`
	ReviewRemark string `json:"review_remark"`
}

func GetOrders(c *gin.Context) {
	userID := c.Query("user_id")
	status := c.Query("status")

	var orders []models.Order
	query := config.DB.Preload("Trip")

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orders})
}

func GetOrder(c *gin.Context) {
	var order models.Order
	if err := config.DB.Preload("Trip").Preload("Trip.Itineraries").Preload("User").First(&order, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": order})
}

func CreateOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if order.TripID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "行程ID不能为空"})
		return
	}

	var trip models.Trip
	if err := config.DB.First(&trip, order.TripID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "行程不存在"})
		return
	}

	if trip.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该行程不可预订"})
		return
	}

	if order.Travelers <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "出行人数必须大于0"})
		return
	}

	if trip.LeftSpots < order.Travelers {
		c.JSON(http.StatusBadRequest, gin.H{"error": "剩余名额不足"})
		return
	}

	if order.ContactName == "" || order.ContactPhone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "联系人和联系电话不能为空"})
		return
	}

	if order.UserID != nil && *order.UserID == 0 {
		order.UserID = nil
	}

	order.OrderNo = generateOrderNo()
	order.Status = "pending"
	order.TripName = trip.Name
	order.TripPrice = trip.Price

	if order.TotalAmount <= 0 {
		order.TotalAmount = trip.Price * float64(order.Travelers)
	}

	tx := config.DB.Begin()

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Model(&trip).Update("left_spots", trip.LeftSpots-order.Travelers).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{"data": order})
}

func PayOrder(c *gin.Context) {
	var order models.Order
	if err := config.DB.First(&order, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	if order.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单状态不允许支付"})
		return
	}

	now := timeNow()
	order.Status = "paid"
	order.PayTime = &now
	config.DB.Save(&order)

	c.JSON(http.StatusOK, gin.H{"data": order})
}

func GetRefundRequests(c *gin.Context) {
	userID := c.Query("user_id")
	status := c.Query("status")

	var requests []models.RefundRequest
	query := config.DB.Preload("Order").Preload("Order.Trip").Preload("User")

	if userID != "" {
		query = query.Where("refund_requests.user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("refund_requests.status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": requests})
}

func GetRefundRequest(c *gin.Context) {
	var req models.RefundRequest
	if err := config.DB.Preload("Order").Preload("Order.Trip").Preload("User").
		Preload("ReviewLogs").Preload("ReviewLogs.Operator").
		First(&req, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "退款申请不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": req})
}

func GetRefundReviewLogs(c *gin.Context) {
	refundID := c.Param("id")
	var logs []models.RefundReviewLog
	if err := config.DB.Preload("Operator").Where("refund_id = ?", refundID).Order("created_at ASC").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func CreateRefundRequest(c *gin.Context) {
	var input RefundRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	if err := config.DB.First(&order, input.OrderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	if order.Status != "paid" && order.Status != "confirmed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前订单状态不支持退款"})
		return
	}

	var existingReq models.RefundRequest
	result := config.DB.Where("order_id = ? AND status IN ?", input.OrderID, []string{"pending", "approved"}).First(&existingReq)
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该订单已有退款申请处理中"})
		return
	}

	refundAmount := input.RefundAmount
	if refundAmount <= 0 || refundAmount > order.TotalAmount {
		refundAmount = order.TotalAmount
	}

	refundReq := models.RefundRequest{
		RefundNo:     generateRefundNo(),
		OrderID:      input.OrderID,
		OrderNo:      order.OrderNo,
		Reason:       input.Reason,
		Description:  input.Description,
		RefundAmount: refundAmount,
		Status:       "pending",
	}
	if order.UserID != nil {
		refundReq.UserID = order.UserID
	}

	if err := config.DB.Create(&refundReq).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	submitLog := models.RefundReviewLog{
		RefundID:   refundReq.ID,
		Action:     "submitted",
		FromStatus: "",
		ToStatus:   "pending",
		Remark:     refundReq.Description,
		OperatorID: refundReq.UserID,
	}
	config.DB.Create(&submitLog)

	order.Status = "refunding"
	config.DB.Save(&order)

	c.JSON(http.StatusCreated, gin.H{"data": refundReq})
}

func ReviewRefundRequest(c *gin.Context) {
	var input ReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status != "approved" && input.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审核状态"})
		return
	}

	var req models.RefundRequest
	if err := config.DB.First(&req, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "退款申请不存在"})
		return
	}

	if req.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该申请已处理，无法重复审核"})
		return
	}

	var order models.Order
	if err := config.DB.Preload("Trip").First(&order, req.OrderID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "关联订单查询失败"})
		return
	}

	tx := config.DB.Begin()

	now := timeNow()
	reviewerID := uint(1)
	req.Status = input.Status
	req.ReviewRemark = input.ReviewRemark
	req.ReviewTime = &now
	req.ReviewerID = &reviewerID
	if err := tx.Save(&req).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	reviewLog := models.RefundReviewLog{
		RefundID:   req.ID,
		Action:     input.Status,
		FromStatus: "pending",
		ToStatus:   input.Status,
		Remark:     input.ReviewRemark,
		OperatorID: &reviewerID,
	}
	if err := tx.Create(&reviewLog).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if input.Status == "approved" {
		order.Status = "refunded"
		var trip models.Trip
		if err := tx.First(&trip, order.TripID).Error; err == nil {
			oldSpots := trip.LeftSpots
			newSpots := trip.LeftSpots + order.Travelers
			if newSpots > trip.TotalSpots {
				newSpots = trip.TotalSpots
			}

			refundID := req.ID
			orderID := order.ID
			adjustLog := models.SpotAdjustmentLog{
				TripID:       trip.ID,
				AdjustType:   "refund_return",
				OldSpots:     oldSpots,
				NewSpots:     newSpots,
				AdjustAmount: order.Travelers,
				Reason:       "退款审核通过，归还余位",
				RefundID:     &refundID,
				OrderID:      &orderID,
			}
			if err := tx.Create(&adjustLog).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			trip.LeftSpots = newSpots
			if err := tx.Save(&trip).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	} else {
		order.Status = "paid"
	}
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"data": req})
}

func BatchReviewRefundRequests(c *gin.Context) {
	var input BatchReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Status != "approved" && input.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审核状态"})
		return
	}

	var requests []models.RefundRequest
	if err := config.DB.Where("id IN ? AND status = 'pending'", input.IDs).Find(&requests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(requests) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有可审核的退款申请"})
		return
	}

	tx := config.DB.Begin()
	now := timeNow()
	reviewerID := uint(1)
	successCount := 0

	for _, req := range requests {
		req.Status = input.Status
		req.ReviewRemark = input.ReviewRemark
		req.ReviewTime = &now
		req.ReviewerID = &reviewerID
		if err := tx.Save(&req).Error; err != nil {
			continue
		}

		batchLog := models.RefundReviewLog{
			RefundID:   req.ID,
			Action:     input.Status,
			FromStatus: "pending",
			ToStatus:   input.Status,
			Remark:     input.ReviewRemark,
			OperatorID: &reviewerID,
		}
		tx.Create(&batchLog)

		var order models.Order
		if err := tx.Preload("Trip").First(&order, req.OrderID).Error; err != nil {
			continue
		}

		if input.Status == "approved" {
			order.Status = "refunded"
			var trip models.Trip
			if err := tx.First(&trip, order.TripID).Error; err == nil {
				oldSpots := trip.LeftSpots
				newSpots := trip.LeftSpots + order.Travelers
				if newSpots > trip.TotalSpots {
					newSpots = trip.TotalSpots
				}

				refundID := req.ID
				orderID := order.ID
				adjustLog := models.SpotAdjustmentLog{
					TripID:       trip.ID,
					AdjustType:   "refund_return_batch",
					OldSpots:     oldSpots,
					NewSpots:     newSpots,
					AdjustAmount: order.Travelers,
					Reason:       "批量退款审核通过，归还余位",
					RefundID:     &refundID,
					OrderID:      &orderID,
				}
				if err := tx.Create(&adjustLog).Error; err == nil {
					trip.LeftSpots = newSpots
					tx.Save(&trip)
				}
			}
		} else {
			order.Status = "paid"
		}
		tx.Save(&order)
		successCount++
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{
		"message":       "批量审核完成",
		"success_count": successCount,
		"total_count":   len(input.IDs),
	})
}

func GetTrips(c *gin.Context) {
	var trips []models.Trip
	if err := config.DB.Order("created_at DESC").Find(&trips).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": trips})
}

func GetTrip(c *gin.Context) {
	var trip models.Trip
	if err := config.DB.Preload("Itineraries").First(&trip, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "行程不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": trip})
}

func CreateTrip(c *gin.Context) {
	var input TripInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := parseDate(input.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式无效"})
		return
	}
	endDate, err := parseDate(input.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式无效"})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期不能早于开始日期"})
		return
	}

	trip := models.Trip{
		Name:        input.Name,
		Description: input.Description,
		Destination: input.Destination,
		StartDate:   startDate,
		EndDate:     endDate,
		Price:       input.Price,
		TotalSpots:  input.TotalSpots,
		LeftSpots:   input.TotalSpots,
		Status:      "active",
	}
	if input.Status != "" {
		trip.Status = input.Status
	}

	tx := config.DB.Begin()
	if err := tx.Create(&trip).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	adjustLog := models.SpotAdjustmentLog{
		TripID:       trip.ID,
		AdjustType:   "init",
		OldSpots:     0,
		NewSpots:     trip.LeftSpots,
		AdjustAmount: trip.TotalSpots,
		Reason:       "行程创建初始化",
	}
	if err := tx.Create(&adjustLog).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusCreated, gin.H{"data": trip})
}

func UpdateTrip(c *gin.Context) {
	var trip models.Trip
	if err := config.DB.First(&trip, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "行程不存在"})
		return
	}

	var input TripInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := parseDate(input.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式无效"})
		return
	}
	endDate, err := parseDate(input.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式无效"})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期不能早于开始日期"})
		return
	}

	oldTotalSpots := trip.TotalSpots
	spotDiff := input.TotalSpots - oldTotalSpots

	tx := config.DB.Begin()

	trip.Name = input.Name
	trip.Description = input.Description
	trip.Destination = input.Destination
	trip.StartDate = startDate
	trip.EndDate = endDate
	trip.Price = input.Price
	trip.TotalSpots = input.TotalSpots

	if spotDiff > 0 {
		trip.LeftSpots += spotDiff
		adjustLog := models.SpotAdjustmentLog{
			TripID:       trip.ID,
			AdjustType:   "increase_total",
			OldSpots:     trip.LeftSpots - spotDiff,
			NewSpots:     trip.LeftSpots,
			AdjustAmount: spotDiff,
			Reason:       "调整总名额增加",
		}
		if err := tx.Create(&adjustLog).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else if spotDiff < 0 {
		usedSpots := trip.TotalSpots - trip.LeftSpots
		if input.TotalSpots < usedSpots {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "总名额不能少于已售出名额"})
			return
		}
		trip.LeftSpots += spotDiff
		adjustLog := models.SpotAdjustmentLog{
			TripID:       trip.ID,
			AdjustType:   "decrease_total",
			OldSpots:     trip.LeftSpots - spotDiff,
			NewSpots:     trip.LeftSpots,
			AdjustAmount: spotDiff,
			Reason:       "调整总名额减少",
		}
		if err := tx.Create(&adjustLog).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if input.Status != "" {
		trip.Status = input.Status
	}

	if err := tx.Save(&trip).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"data": trip})
}

func DeleteTrip(c *gin.Context) {
	var trip models.Trip
	if err := config.DB.First(&trip, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "行程不存在"})
		return
	}

	var orderCount int64
	config.DB.Model(&models.Order{}).Where("trip_id = ?", trip.ID).Count(&orderCount)
	if orderCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该行程已有订单，无法删除"})
		return
	}

	tx := config.DB.Begin()
	if err := tx.Where("trip_id = ?", trip.ID).Delete(&models.TripItinerary{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Where("trip_id = ?", trip.ID).Delete(&models.SpotAdjustmentLog{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Delete(&trip).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func AdjustTripSpots(c *gin.Context) {
	var trip models.Trip
	if err := config.DB.First(&trip, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "行程不存在"})
		return
	}

	var input SpotAdjustInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.AdjustAmount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "调整数量不能为0"})
		return
	}

	newLeftSpots := trip.LeftSpots + input.AdjustAmount
	if newLeftSpots < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "调整后余位不能为负数"})
		return
	}
	if newLeftSpots > trip.TotalSpots {
		c.JSON(http.StatusBadRequest, gin.H{"error": "调整后余位不能超过总名额"})
		return
	}

	adjustType := "manual_increase"
	if input.AdjustAmount < 0 {
		adjustType = "manual_decrease"
	}

	adjustLog := models.SpotAdjustmentLog{
		TripID:       trip.ID,
		AdjustType:   adjustType,
		OldSpots:     trip.LeftSpots,
		NewSpots:     newLeftSpots,
		AdjustAmount: input.AdjustAmount,
		Reason:       input.Reason,
	}

	tx := config.DB.Begin()
	trip.LeftSpots = newLeftSpots
	if err := tx.Save(&trip).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Create(&adjustLog).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"data": trip, "log": adjustLog})
}

func GetTripItineraries(c *gin.Context) {
	tripID := c.Param("id")
	var itineraries []models.TripItinerary
	if err := config.DB.Where("trip_id = ?", tripID).Order("day_number ASC").Find(&itineraries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": itineraries})
}

func CreateItinerary(c *gin.Context) {
	tripID := c.Param("id")
	var trip models.Trip
	if err := config.DB.First(&trip, tripID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "行程不存在"})
		return
	}

	var input ItineraryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.DayNumber < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "天数必须大于0"})
		return
	}

	var existing models.TripItinerary
	result := config.DB.Where("trip_id = ? AND day_number = ?", tripID, input.DayNumber).First(&existing)
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该天行程已存在"})
		return
	}

	itinerary := models.TripItinerary{
		TripID:         trip.ID,
		DayNumber:      input.DayNumber,
		Title:          input.Title,
		Breakfast:      input.Breakfast,
		Lunch:          input.Lunch,
		Dinner:         input.Dinner,
		Accommodation:  input.Accommodation,
		Transportation: input.Transportation,
		Activities:     input.Activities,
		Notes:          input.Notes,
	}

	if err := config.DB.Create(&itinerary).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": itinerary})
}

func UpdateItinerary(c *gin.Context) {
	var itinerary models.TripItinerary
	if err := config.DB.First(&itinerary, c.Param("itinerary_id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "行程安排不存在"})
		return
	}

	var input ItineraryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.DayNumber < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "天数必须大于0"})
		return
	}

	if input.DayNumber != itinerary.DayNumber {
		var existing models.TripItinerary
		result := config.DB.Where("trip_id = ? AND day_number = ? AND id != ?", itinerary.TripID, input.DayNumber, itinerary.ID).First(&existing)
		if result.Error == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "该天行程已存在"})
			return
		}
	}

	itinerary.DayNumber = input.DayNumber
	itinerary.Title = input.Title
	itinerary.Breakfast = input.Breakfast
	itinerary.Lunch = input.Lunch
	itinerary.Dinner = input.Dinner
	itinerary.Accommodation = input.Accommodation
	itinerary.Transportation = input.Transportation
	itinerary.Activities = input.Activities
	itinerary.Notes = input.Notes

	if err := config.DB.Save(&itinerary).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": itinerary})
}

func DeleteItinerary(c *gin.Context) {
	var itinerary models.TripItinerary
	if err := config.DB.First(&itinerary, c.Param("itinerary_id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "行程安排不存在"})
		return
	}

	if err := config.DB.Delete(&itinerary).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func GetSpotLogs(c *gin.Context) {
	tripID := c.Param("id")
	var logs []models.SpotAdjustmentLog
	if err := config.DB.Where("trip_id = ?", tripID).Order("created_at DESC").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func GetUsers(c *gin.Context) {
	var users []models.User
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}

func SeedData(c *gin.Context) {
	users := []models.User{
		{Username: "admin", Email: "admin@test.com", Phone: "13800000000", Role: "admin"},
		{Username: "user1", Email: "user1@test.com", Phone: "13800000001", Role: "user"},
		{Username: "user2", Email: "user2@test.com", Phone: "13800000002", Role: "user"},
	}
	for _, u := range users {
		config.DB.FirstOrCreate(&u, models.User{Username: u.Username})
	}

	now := timeNow()
	trips := []models.Trip{
		{Name: "云南大理5日游", Description: "探索风花雪月的大理古城", Destination: "云南大理",
			StartDate: now.AddDate(0, 1, 0), EndDate: now.AddDate(0, 1, 4), Price: 2999, TotalSpots: 20, LeftSpots: 15, Status: "active"},
		{Name: "三亚海岛度假游", Description: "阳光沙滩，椰林海风", Destination: "海南三亚",
			StartDate: now.AddDate(0, 2, 0), EndDate: now.AddDate(0, 2, 6), Price: 4599, TotalSpots: 30, LeftSpots: 25, Status: "active"},
		{Name: "西藏拉萨朝圣之旅", Description: "雪域高原，心灵之旅", Destination: "西藏拉萨",
			StartDate: now.AddDate(0, 3, 0), EndDate: now.AddDate(0, 3, 7), Price: 6899, TotalSpots: 15, LeftSpots: 10, Status: "active"},
	}
	for _, t := range trips {
		config.DB.FirstOrCreate(&t, models.Trip{Name: t.Name})
	}

	daliItineraries := []models.TripItinerary{
		{TripID: 1, DayNumber: 1, Title: "出发抵达大理", Breakfast: "自理", Lunch: "自理", Dinner: "大理特色白族菜",
			Accommodation: "大理古城特色客栈", Transportation: "飞机+专车接送",
			Activities: "抵达大理机场，专车接往古城，自由活动逛古城夜景", Notes: "建议携带防晒用品"},
		{TripID: 1, DayNumber: 2, Title: "洱海环游", Breakfast: "酒店早餐", Lunch: "洱海渔村特色菜", Dinner: "双廊古镇美食",
			Accommodation: "双廊海景酒店", Transportation: "商务车环湖",
			Activities: "洱海骑行、喜洲古镇、双廊艺术小镇", Notes: "注意防晒，做好高原反应预防"},
		{TripID: 1, DayNumber: 3, Title: "苍山索道+古城深度游", Breakfast: "酒店早餐", Lunch: "大理风味餐厅", Dinner: "特色过桥米线",
			Accommodation: "大理古城特色客栈", Transportation: "商务车+索道",
			Activities: "苍山洗马潭索道、天龙八部影视城、洋人街", Notes: "山顶温度较低，建议携带外套"},
		{TripID: 1, DayNumber: 4, Title: "崇圣寺三塔+返程", Breakfast: "酒店早餐", Lunch: "景区餐厅", Dinner: "自理",
			Accommodation: "无", Transportation: "商务车+飞机",
			Activities: "崇圣寺三塔文化旅游区，送机返程", Notes: "退房时间中午12点前"},
	}
	for _, it := range daliItineraries {
		config.DB.FirstOrCreate(&it, models.TripItinerary{TripID: it.TripID, DayNumber: it.DayNumber})
	}

	sanyaItineraries := []models.TripItinerary{
		{TripID: 2, DayNumber: 1, Title: "抵达三亚，入住酒店", Breakfast: "自理", Lunch: "自理", Dinner: "海鲜大餐",
			Accommodation: "三亚湾海景酒店", Transportation: "飞机+专车接送",
			Activities: "抵达三亚凤凰机场，入住酒店，自由活动", Notes: "做好防晒措施"},
		{TripID: 2, DayNumber: 2, Title: "蜈支洲岛一日游", Breakfast: "酒店早餐", Lunch: "岛上海鲜餐厅", Dinner: "沙滩BBQ",
			Accommodation: "三亚湾海景酒店", Transportation: "游船+专车",
			Activities: "蜈支洲岛观光、海上项目可选（潜水、摩托艇等）", Notes: "海上项目需自费"},
		{TripID: 2, DayNumber: 3, Title: "热带雨林+南山寺", Breakfast: "酒店早餐", Lunch: "素斋", Dinner: "椰林特色餐",
			Accommodation: "三亚湾海景酒店", Transportation: "商务车",
			Activities: "呀诺达热带雨林、南山文化旅游区、108米海上观音", Notes: "建议穿舒适运动鞋"},
		{TripID: 2, DayNumber: 4, Title: "免税购物+返程", Breakfast: "酒店早餐", Lunch: "特色小吃", Dinner: "自理",
			Accommodation: "无", Transportation: "专车+飞机",
			Activities: "三亚国际免税城购物，送机返程", Notes: "记得带身份证购物"},
	}
	for _, it := range sanyaItineraries {
		config.DB.FirstOrCreate(&it, models.TripItinerary{TripID: it.TripID, DayNumber: it.DayNumber})
	}

	lhasaItineraries := []models.TripItinerary{
		{TripID: 3, DayNumber: 1, Title: "抵达拉萨，适应高原", Breakfast: "自理", Lunch: "藏式特色餐", Dinner: "酒店餐厅",
			Accommodation: "拉萨市区供氧酒店", Transportation: "飞机+专车接送",
			Activities: "抵达拉萨，酒店休息适应高原环境", Notes: "不要剧烈运动，多喝温水"},
		{TripID: 3, DayNumber: 2, Title: "布达拉宫+大昭寺", Breakfast: "酒店早餐", Lunch: "藏式餐厅", Dinner: "特色小吃",
			Accommodation: "拉萨市区供氧酒店", Transportation: "商务车",
			Activities: "布达拉宫参观、大昭寺转经、八廓街逛", Notes: "参观布达拉宫需提前预约"},
		{TripID: 3, DayNumber: 3, Title: "羊卓雍措一日游", Breakfast: "酒店早餐", Lunch: "景区餐厅", Dinner: "酒店餐厅",
			Accommodation: "拉萨市区供氧酒店", Transportation: "商务车",
			Activities: "圣湖羊卓雍措、卡若拉冰川", Notes: "海拔较高，注意保暖和高反"},
		{TripID: 3, DayNumber: 4, Title: "纳木措一日游", Breakfast: "酒店早餐", Lunch: "简餐", Dinner: "特色餐",
			Accommodation: "拉萨市区供氧酒店", Transportation: "商务车",
			Activities: "天湖纳木措、那根拉山口", Notes: "路程较远，建议备晕车药"},
		{TripID: 3, DayNumber: 5, Title: "返程", Breakfast: "酒店早餐", Lunch: "自理", Dinner: "自理",
			Accommodation: "无", Transportation: "专车+飞机",
			Activities: "自由活动，送机返程", Notes: "检查好随身物品"},
	}
	for _, it := range lhasaItineraries {
		config.DB.FirstOrCreate(&it, models.TripItinerary{TripID: it.TripID, DayNumber: it.DayNumber})
	}

	c.JSON(http.StatusOK, gin.H{"message": "测试数据初始化成功"})
}
