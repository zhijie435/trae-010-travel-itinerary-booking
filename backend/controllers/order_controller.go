package controllers

import (
	"net/http"
	"strconv"
	"travel-refund/models"
	"travel-refund/services"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	orderService *services.OrderService
}

func NewOrderController() *OrderController {
	return &OrderController{
		orderService: services.NewOrderService(),
	}
}

type RefundRequestInput struct {
	OrderID      uint    `json:"order_id" binding:"required"`
	Reason       string  `json:"reason" binding:"required"`
	Description  string  `json:"description"`
	RefundAmount float64 `json:"refund_amount"`
}

type ReviewInput struct {
	Status       string `json:"status" binding:"required"`
	ReviewRemark string `json:"review_remark"`
}

type BatchReviewInput struct {
	IDs          []uint `json:"ids" binding:"required,min=1"`
	Status       string `json:"status" binding:"required"`
	ReviewRemark string `json:"review_remark"`
}

func (ctrl *OrderController) GetOrders(c *gin.Context) {
	userID := c.Query("user_id")
	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	orders, err := ctrl.orderService.List(userID, status, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": orders})
}

func (ctrl *OrderController) GetOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	order, err := ctrl.orderService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

func (ctrl *OrderController) CreateOrder(c *gin.Context) {
	var order models.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if order.RouteID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "线路ID不能为空"})
		return
	}

	newOrder, err := ctrl.orderService.Create(
		order.RouteID, order.Travelers,
		order.ContactName, order.ContactPhone, order.UserID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if newOrder == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单创建失败，请检查参数"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": newOrder})
}

func (ctrl *OrderController) PayOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	order, err := ctrl.orderService.Pay(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}
	if order == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单状态不允许支付"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

func (ctrl *OrderController) GetRefundRequests(c *gin.Context) {
	userID := c.Query("user_id")
	status := c.Query("status")

	requests, err := ctrl.orderService.ListRefunds(userID, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": requests})
}

func (ctrl *OrderController) GetRefundRequest(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	req, err := ctrl.orderService.GetRefundByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "退款申请不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": req})
}

func (ctrl *OrderController) GetRefundReviewLogs(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	logs, err := ctrl.orderService.GetRefundReviewLogs(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}

func (ctrl *OrderController) CreateRefundRequest(c *gin.Context) {
	var input RefundRequestInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req, err := ctrl.orderService.CreateRefund(
		input.OrderID, input.Reason, input.Description, input.RefundAmount,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "当前订单状态不支持退款或已有处理中的申请"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": req})
}

func (ctrl *OrderController) ReviewRefundRequest(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var input ReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req, err := ctrl.orderService.ReviewRefund(uint(id), input.Status, input.ReviewRemark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if req == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的审核状态或申请已处理"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": req})
}

func (ctrl *OrderController) BatchReviewRefundRequests(c *gin.Context) {
	var input BatchReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	successCount, err := ctrl.orderService.BatchReviewRefunds(input.IDs, input.Status, input.ReviewRemark)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "批量审核完成",
		"success_count": successCount,
		"total_count":   len(input.IDs),
	})
}
