package services

import (
	"travel-refund/config"
	"travel-refund/models"
	"travel-refund/pkg/utils"
	"time"
)

type OrderService struct{}

func NewOrderService() *OrderService {
	return &OrderService{}
}

func (s *OrderService) List(userID, status, startDate, endDate string) ([]models.Order, error) {
	var orders []models.Order
	query := config.DB.Preload("Route").Preload("Inventory")

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if startDate != "" {
		if t, err := utils.ParseDate(startDate); err == nil {
			query = query.Where("created_at >= ?", t)
		}
	}
	if endDate != "" {
		if t, err := utils.ParseDate(endDate); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			query = query.Where("created_at <= ?", endOfDay)
		}
	}

	if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (s *OrderService) GetByID(id uint) (*models.Order, error) {
	var order models.Order
	if err := config.DB.Preload("Route").Preload("Route.Itineraries").Preload("Inventory").Preload("User").First(&order, id).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

func (s *OrderService) Create(routeID uint, travelers int, contactName, contactPhone string, userID *uint) (*models.Order, error) {
	var route models.Route
	if err := config.DB.First(&route, routeID).Error; err != nil {
		return nil, err
	}

	if route.Status != "active" {
		return nil, nil
	}

	if travelers <= 0 {
		return nil, nil
	}

	inventory, err := NewInventoryService().GetByRouteAndDate(routeID, route.StartDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}

	if inventory.LeftSpots < travelers {
		return nil, nil
	}

	if contactName == "" || contactPhone == "" {
		return nil, nil
	}

	if userID != nil && *userID == 0 {
		userID = nil
	}

	order := &models.Order{
		OrderNo:      utils.GenerateOrderNo(),
		UserID:       userID,
		RouteID:      routeID,
		InventoryID:  inventory.ID,
		RouteName:    route.Name,
		RoutePrice:   route.Price,
		Travelers:    travelers,
		TotalAmount:  route.Price * float64(travelers),
		ContactName:  contactName,
		ContactPhone: contactPhone,
		Status:       "pending",
	}

	tx := config.DB.Begin()

	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	inventoryService := NewInventoryService()
	_, err = inventoryService.AdjustWithOrder(tx, inventory.ID, -travelers, "order_lock", "订单预订锁定名额", order.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return order, nil
}

func (s *OrderService) Pay(id uint) (*models.Order, error) {
	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		return nil, err
	}

	if order.Status != "pending" {
		return nil, nil
	}

	now := utils.TimeNow()
	order.Status = "paid"
	order.PayTime = &now
	config.DB.Save(&order)

	return &order, nil
}

func (s *OrderService) Cancel(id uint) (*models.Order, error) {
	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		return nil, err
	}

	if order.Status != "pending" {
		return nil, nil
	}

	tx := config.DB.Begin()

	order.Status = "cancelled"
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	inventoryService := NewInventoryService()
	_, err := inventoryService.AdjustWithOrder(tx, order.InventoryID, order.Travelers, "order_cancel", "订单取消释放名额", order.ID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return &order, nil
}

func (s *OrderService) CreateRefund(orderID uint, reason, description string, refundAmount float64) (*models.RefundRequest, error) {
	var order models.Order
	if err := config.DB.First(&order, orderID).Error; err != nil {
		return nil, err
	}

	if order.Status != "paid" && order.Status != "confirmed" {
		return nil, nil
	}

	var existingReq models.RefundRequest
	result := config.DB.Where("order_id = ? AND status IN ?", orderID, []string{"pending", "approved"}).First(&existingReq)
	if result.Error == nil {
		return nil, nil
	}

	if refundAmount <= 0 || refundAmount > order.TotalAmount {
		refundAmount = order.TotalAmount
	}

	refundReq := &models.RefundRequest{
		RefundNo:     utils.GenerateRefundNo(),
		OrderID:      orderID,
		OrderNo:      order.OrderNo,
		Reason:       reason,
		Description:  description,
		RefundAmount: refundAmount,
		Status:       "pending",
	}
	if order.UserID != nil {
		refundReq.UserID = order.UserID
	}

	tx := config.DB.Begin()

	if err := tx.Create(refundReq).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	submitLog := &models.RefundReviewLog{
		RefundID:   refundReq.ID,
		Action:     "submitted",
		FromStatus: "",
		ToStatus:   "pending",
		Remark:     refundReq.Description,
		OperatorID: refundReq.UserID,
	}
	if err := tx.Create(submitLog).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	order.Status = "refunding"
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return refundReq, nil
}

func (s *OrderService) ReviewRefund(id uint, status, reviewRemark string) (*models.RefundRequest, error) {
	if status != "approved" && status != "rejected" {
		return nil, nil
	}

	var req models.RefundRequest
	if err := config.DB.First(&req, id).Error; err != nil {
		return nil, err
	}

	if req.Status != "pending" {
		return nil, nil
	}

	var order models.Order
	if err := config.DB.Preload("Route").Preload("Inventory").First(&order, req.OrderID).Error; err != nil {
		return nil, err
	}

	tx := config.DB.Begin()

	now := utils.TimeNow()
	reviewerID := uint(1)
	req.Status = status
	req.ReviewRemark = reviewRemark
	req.ReviewTime = &now
	req.ReviewerID = &reviewerID
	if err := tx.Save(&req).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	reviewLog := &models.RefundReviewLog{
		RefundID:   req.ID,
		Action:     status,
		FromStatus: "pending",
		ToStatus:   status,
		Remark:     reviewRemark,
		OperatorID: &reviewerID,
	}
	if err := tx.Create(reviewLog).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if status == "approved" {
		order.Status = "refunded"
		inventoryService := NewInventoryService()
		_, err := inventoryService.AdjustWithRefund(tx, order.InventoryID, order.Travelers, "refund_return", "退款审核通过，归还余位", req.ID, order.ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		order.Status = "paid"
	}
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return &req, nil
}

func (s *OrderService) BatchReviewRefunds(ids []uint, status, reviewRemark string) (int, error) {
	if status != "approved" && status != "rejected" {
		return 0, nil
	}

	var requests []models.RefundRequest
	if err := config.DB.Where("id IN ? AND status = 'pending'", ids).Find(&requests).Error; err != nil {
		return 0, err
	}

	if len(requests) == 0 {
		return 0, nil
	}

	tx := config.DB.Begin()
	now := utils.TimeNow()
	reviewerID := uint(1)
	successCount := 0

	for _, req := range requests {
		req.Status = status
		req.ReviewRemark = reviewRemark
		req.ReviewTime = &now
		req.ReviewerID = &reviewerID
		if err := tx.Save(&req).Error; err != nil {
			continue
		}

		batchLog := &models.RefundReviewLog{
			RefundID:   req.ID,
			Action:     status,
			FromStatus: "pending",
			ToStatus:   status,
			Remark:     reviewRemark,
			OperatorID: &reviewerID,
		}
		tx.Create(batchLog)

		var order models.Order
		if err := tx.Preload("Route").Preload("Inventory").First(&order, req.OrderID).Error; err != nil {
			continue
		}

		if status == "approved" {
			order.Status = "refunded"
			inventoryService := NewInventoryService()
			_, err := inventoryService.AdjustWithRefund(tx, order.InventoryID, order.Travelers, "refund_return_batch", "批量退款审核通过，归还余位", req.ID, order.ID)
			if err == nil {
				tx.Save(&order)
				successCount++
			}
		} else {
			order.Status = "paid"
			tx.Save(&order)
			successCount++
		}
	}

	tx.Commit()
	return successCount, nil
}

func (s *OrderService) ListRefunds(userID, status string) ([]models.RefundRequest, error) {
	var requests []models.RefundRequest
	query := config.DB.Preload("Order").Preload("Order.Route").Preload("Order.Inventory").Preload("User")

	if userID != "" {
		query = query.Where("refund_requests.user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("refund_requests.status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&requests).Error; err != nil {
		return nil, err
	}
	return requests, nil
}

func (s *OrderService) GetRefundByID(id uint) (*models.RefundRequest, error) {
	var req models.RefundRequest
	if err := config.DB.Preload("Order").Preload("Order.Route").Preload("Order.Inventory").Preload("User").
		Preload("ReviewLogs").Preload("ReviewLogs.Operator").
		First(&req, id).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (s *OrderService) GetRefundReviewLogs(refundID uint) ([]models.RefundReviewLog, error) {
	var logs []models.RefundReviewLog
	if err := config.DB.Preload("Operator").Where("refund_id = ?", refundID).Order("created_at ASC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}
