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
	if err := config.DB.Preload("Trip").Preload("User").First(&order, c.Param("id")).Error; err != nil {
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

	order.OrderNo = generateOrderNo()
	order.Status = "pending"

	if err := config.DB.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
	if err := config.DB.Preload("Order").Preload("Order.Trip").Preload("User").First(&req, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "退款申请不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": req})
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
		UserID:       order.UserID,
		Reason:       input.Reason,
		Description:  input.Description,
		RefundAmount: refundAmount,
		Status:       "pending",
	}

	if err := config.DB.Create(&refundReq).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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
	config.DB.First(&order, req.OrderID)

	now := timeNow()
	reviewerID := uint(1)
	req.Status = input.Status
	req.ReviewRemark = input.ReviewRemark
	req.ReviewTime = &now
	req.ReviewerID = &reviewerID
	config.DB.Save(&req)

	if input.Status == "approved" {
		order.Status = "refunded"
	} else {
		order.Status = "paid"
	}
	config.DB.Save(&order)

	c.JSON(http.StatusOK, gin.H{"data": req})
}

func GetTrips(c *gin.Context) {
	var trips []models.Trip
	if err := config.DB.Order("created_at DESC").Find(&trips).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": trips})
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

	c.JSON(http.StatusOK, gin.H{"message": "测试数据初始化成功"})
}
