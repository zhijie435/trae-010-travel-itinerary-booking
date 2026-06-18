package services

import (
	"testing"
	"time"
	"travel-refund/config"
	"travel-refund/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(1)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.Route{},
		&models.Itinerary{},
		&models.Inventory{},
		&models.InventoryAdjustLog{},
		&models.Order{},
		&models.RefundRequest{},
		&models.RefundReviewLog{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	config.DB = db
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		config.DB = nil
	})
}

func seedRouteAndInventory(t *testing.T, totalSpots, leftSpots int, routeStatus string) (*models.Route, *models.Inventory) {
	t.Helper()
	start := time.Date(2026, 7, 1, 9, 0, 0, 0, time.Local)
	route := &models.Route{
		Name:        "测试线路",
		Destination: "测试目的地",
		StartDate:   start,
		EndDate:     start.AddDate(0, 0, 4),
		Price:       1999,
		Status:      routeStatus,
	}
	if err := config.DB.Create(route).Error; err != nil {
		t.Fatalf("create route: %v", err)
	}
	inventory := &models.Inventory{
		RouteID:    route.ID,
		Date:       route.StartDate.Format("2006-01-02"),
		TotalSpots: totalSpots,
		LeftSpots:  leftSpots,
		Status:     "active",
	}
	if err := config.DB.Create(inventory).Error; err != nil {
		t.Fatalf("create inventory: %v", err)
	}
	return route, inventory
}

func reloadInventory(t *testing.T, id uint) *models.Inventory {
	t.Helper()
	var inv models.Inventory
	if err := config.DB.First(&inv, id).Error; err != nil {
		t.Fatalf("reload inventory %d: %v", id, err)
	}
	return &inv
}

func countAdjustLogs(inventoryID uint, adjustType string) int64 {
	var n int64
	config.DB.Model(&models.InventoryAdjustLog{}).
		Where("inventory_id = ? AND adjust_type = ?", inventoryID, adjustType).
		Count(&n)
	return n
}

func latestAdjustLog(inventoryID uint, adjustType string) models.InventoryAdjustLog {
	var log models.InventoryAdjustLog
	config.DB.Where("inventory_id = ? AND adjust_type = ?", inventoryID, adjustType).
		Order("id DESC").First(&log)
	return log
}

func TestCreateOrderLocksInventory(t *testing.T) {
	setupTestDB(t)
	route, inv := seedRouteAndInventory(t, 10, 10, "active")

	svc := NewOrderService()
	order, err := svc.Create(route.ID, 3, "张三", "13800000000", nil)
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if order == nil {
		t.Fatal("expected order created, got nil")
	}
	if order.Status != "pending" {
		t.Errorf("status = %q, want pending", order.Status)
	}
	if order.Travelers != 3 {
		t.Errorf("travelers = %d, want 3", order.Travelers)
	}
	if order.InventoryID != inv.ID {
		t.Errorf("inventory_id = %d, want %d", order.InventoryID, inv.ID)
	}
	if order.TotalAmount != route.Price*3 {
		t.Errorf("total_amount = %v, want %v", order.TotalAmount, route.Price*3)
	}

	got := reloadInventory(t, inv.ID)
	if got.LeftSpots != 7 {
		t.Errorf("left_spots = %d, want 7 (10-3)", got.LeftSpots)
	}

	if n := countAdjustLogs(inv.ID, "order_lock"); n != 1 {
		t.Errorf("order_lock logs = %d, want 1", n)
	}
	log := latestAdjustLog(inv.ID, "order_lock")
	if log.AdjustAmount != -3 {
		t.Errorf("lock log adjust_amount = %d, want -3", log.AdjustAmount)
	}
	if log.OldSpots != 10 || log.NewSpots != 7 {
		t.Errorf("lock log old=%d new=%d, want old=10 new=7", log.OldSpots, log.NewSpots)
	}
	if log.OrderID == nil || *log.OrderID != order.ID {
		t.Errorf("lock log order_id = %v, want %d", log.OrderID, order.ID)
	}
}

func TestCancelOrderReleasesInventory(t *testing.T) {
	setupTestDB(t)
	route, inv := seedRouteAndInventory(t, 10, 10, "active")

	svc := NewOrderService()
	order, err := svc.Create(route.ID, 3, "张三", "13800000000", nil)
	if err != nil || order == nil {
		t.Fatalf("create order: err=%v order=%v", err, order)
	}

	cancelled, err := svc.Cancel(order.ID)
	if err != nil {
		t.Fatalf("cancel order: %v", err)
	}
	if cancelled == nil {
		t.Fatal("expected order cancelled, got nil")
	}
	if cancelled.Status != "cancelled" {
		t.Errorf("status = %q, want cancelled", cancelled.Status)
	}

	got := reloadInventory(t, inv.ID)
	if got.LeftSpots != 10 {
		t.Errorf("left_spots = %d, want 10 (released)", got.LeftSpots)
	}

	if n := countAdjustLogs(inv.ID, "order_cancel"); n != 1 {
		t.Errorf("order_cancel logs = %d, want 1", n)
	}
	log := latestAdjustLog(inv.ID, "order_cancel")
	if log.AdjustAmount != 3 {
		t.Errorf("cancel log adjust_amount = %d, want 3", log.AdjustAmount)
	}
	if log.OldSpots != 7 || log.NewSpots != 10 {
		t.Errorf("cancel log old=%d new=%d, want old=7 new=10", log.OldSpots, log.NewSpots)
	}
	if log.OrderID == nil || *log.OrderID != order.ID {
		t.Errorf("cancel log order_id = %v, want %d", log.OrderID, order.ID)
	}
}

func TestCreateOrderFailsWhenInsufficientSpots(t *testing.T) {
	setupTestDB(t)
	route, inv := seedRouteAndInventory(t, 10, 2, "active")

	svc := NewOrderService()
	order, err := svc.Create(route.ID, 3, "张三", "13800000000", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order != nil {
		t.Errorf("expected nil order when not enough spots, got %+v", order)
	}

	got := reloadInventory(t, inv.ID)
	if got.LeftSpots != 2 {
		t.Errorf("left_spots = %d, want 2 (unchanged)", got.LeftSpots)
	}
	if n := countAdjustLogs(inv.ID, "order_lock"); n != 0 {
		t.Errorf("order_lock logs = %d, want 0", n)
	}
	var orderCount int64
	config.DB.Model(&models.Order{}).Count(&orderCount)
	if orderCount != 0 {
		t.Errorf("orders = %d, want 0 (no order persisted)", orderCount)
	}
}

func TestCreateOrderFailsWhenRouteInactive(t *testing.T) {
	setupTestDB(t)
	route, _ := seedRouteAndInventory(t, 10, 10, "inactive")

	svc := NewOrderService()
	order, err := svc.Create(route.ID, 2, "张三", "13800000000", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order != nil {
		t.Errorf("expected nil order for inactive route, got %+v", order)
	}
	var orderCount int64
	config.DB.Model(&models.Order{}).Count(&orderCount)
	if orderCount != 0 {
		t.Errorf("orders = %d, want 0", orderCount)
	}
}

func TestCancelOrderTwiceDoesNotDoubleRelease(t *testing.T) {
	setupTestDB(t)
	route, inv := seedRouteAndInventory(t, 10, 10, "active")

	svc := NewOrderService()
	order, err := svc.Create(route.ID, 3, "张三", "13800000000", nil)
	if err != nil || order == nil {
		t.Fatalf("create order: err=%v order=%v", err, order)
	}

	if _, err := svc.Cancel(order.ID); err != nil {
		t.Fatalf("first cancel: %v", err)
	}

	second, err := svc.Cancel(order.ID)
	if err != nil {
		t.Fatalf("second cancel error: %v", err)
	}
	if second != nil {
		t.Errorf("expected nil on second cancel, got %+v", second)
	}

	got := reloadInventory(t, inv.ID)
	if got.LeftSpots != 10 {
		t.Errorf("left_spots = %d, want 10 (no double release)", got.LeftSpots)
	}
	if n := countAdjustLogs(inv.ID, "order_cancel"); n != 1 {
		t.Errorf("order_cancel logs = %d, want 1", n)
	}
}

func TestRefundApprovedReleasesInventory(t *testing.T) {
	setupTestDB(t)
	route, inv := seedRouteAndInventory(t, 10, 10, "active")

	svc := NewOrderService()
	order, err := svc.Create(route.ID, 3, "张三", "13800000000", nil)
	if err != nil || order == nil {
		t.Fatalf("create order: err=%v order=%v", err, order)
	}

	if _, err := svc.Pay(order.ID); err != nil {
		t.Fatalf("pay: %v", err)
	}
	if got := reloadInventory(t, inv.ID); got.LeftSpots != 7 {
		t.Fatalf("after pay left_spots = %d, want 7 (pay must not touch inventory)", got.LeftSpots)
	}

	refund, err := svc.CreateRefund(order.ID, "行程冲突", "临时有事无法出行", 0)
	if err != nil || refund == nil {
		t.Fatalf("create refund: err=%v refund=%v", err, refund)
	}

	if _, err := svc.ReviewRefund(refund.ID, "approved", "同意退款"); err != nil {
		t.Fatalf("review refund: %v", err)
	}

	got := reloadInventory(t, inv.ID)
	if got.LeftSpots != 10 {
		t.Errorf("left_spots = %d, want 10 (refund released)", got.LeftSpots)
	}
	if n := countAdjustLogs(inv.ID, "refund_return"); n != 1 {
		t.Errorf("refund_return logs = %d, want 1", n)
	}
	log := latestAdjustLog(inv.ID, "refund_return")
	if log.AdjustAmount != 3 {
		t.Errorf("refund log adjust_amount = %d, want 3", log.AdjustAmount)
	}

	var dbOrder models.Order
	config.DB.First(&dbOrder, order.ID)
	if dbOrder.Status != "refunded" {
		t.Errorf("order status = %q, want refunded", dbOrder.Status)
	}
}
