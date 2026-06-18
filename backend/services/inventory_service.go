package services

import (
	"travel-refund/config"
	"travel-refund/models"

	"gorm.io/gorm"
)

type InventoryService struct{}

func NewInventoryService() *InventoryService {
	return &InventoryService{}
}

func (s *InventoryService) GetByID(id uint) (*models.Inventory, error) {
	var inventory models.Inventory
	if err := config.DB.Preload("Route").First(&inventory, id).Error; err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (s *InventoryService) GetByRouteAndDate(routeID uint, date string) (*models.Inventory, error) {
	var inventory models.Inventory
	if err := config.DB.Where("route_id = ? AND date = ?", routeID, date).First(&inventory).Error; err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (s *InventoryService) ListByRoute(routeID uint) ([]models.Inventory, error) {
	var inventories []models.Inventory
	if err := config.DB.Where("route_id = ?", routeID).Order("date ASC").Find(&inventories).Error; err != nil {
		return nil, err
	}
	return inventories, nil
}

func (s *InventoryService) Create(routeID uint, date string, totalSpots int, adjustType, reason string) (*models.Inventory, error) {
	tx := config.DB.Begin()
	inventory, err := s.CreateWithTx(tx, routeID, date, totalSpots, adjustType, reason)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return inventory, nil
}

func (s *InventoryService) CreateWithTx(tx *gorm.DB, routeID uint, date string, totalSpots int, adjustType, reason string) (*models.Inventory, error) {
	inventory := &models.Inventory{
		RouteID:    routeID,
		Date:       date,
		TotalSpots: totalSpots,
		LeftSpots:  totalSpots,
		Status:     "active",
	}

	if err := tx.Create(inventory).Error; err != nil {
		return nil, err
	}

	adjustLog := &models.InventoryAdjustLog{
		InventoryID:  inventory.ID,
		RouteID:      routeID,
		AdjustType:   adjustType,
		OldSpots:     0,
		NewSpots:     inventory.LeftSpots,
		AdjustAmount: totalSpots,
		Reason:       reason,
	}

	if err := tx.Create(adjustLog).Error; err != nil {
		return nil, err
	}

	return inventory, nil
}

func (s *InventoryService) Adjust(inventoryID uint, adjustAmount int, adjustType, reason string) (*models.Inventory, error) {
	tx := config.DB.Begin()
	inventory, err := s.AdjustWithTx(tx, inventoryID, adjustAmount, adjustType, reason)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return inventory, nil
}

func (s *InventoryService) AdjustWithTx(tx *gorm.DB, inventoryID uint, adjustAmount int, adjustType, reason string) (*models.Inventory, error) {
	var inventory models.Inventory
	if err := tx.First(&inventory, inventoryID).Error; err != nil {
		return nil, err
	}

	newLeftSpots := inventory.LeftSpots + adjustAmount
	if newLeftSpots < 0 {
		return nil, nil
	}
	if newLeftSpots > inventory.TotalSpots {
		return nil, nil
	}

	adjustLog := &models.InventoryAdjustLog{
		InventoryID:  inventoryID,
		RouteID:      inventory.RouteID,
		AdjustType:   adjustType,
		OldSpots:     inventory.LeftSpots,
		NewSpots:     newLeftSpots,
		AdjustAmount: adjustAmount,
		Reason:       reason,
	}

	inventory.LeftSpots = newLeftSpots
	if err := tx.Save(&inventory).Error; err != nil {
		return nil, err
	}

	if err := tx.Create(adjustLog).Error; err != nil {
		return nil, err
	}

	return &inventory, nil
}

func (s *InventoryService) AdjustWithOrder(tx *gorm.DB, inventoryID uint, adjustAmount int, adjustType, reason string, orderID uint) (*models.Inventory, error) {
	var inventory models.Inventory
	if err := tx.First(&inventory, inventoryID).Error; err != nil {
		return nil, err
	}

	newLeftSpots := inventory.LeftSpots + adjustAmount
	if newLeftSpots < 0 {
		return nil, nil
	}
	if newLeftSpots > inventory.TotalSpots {
		return nil, nil
	}

	adjustLog := &models.InventoryAdjustLog{
		InventoryID:  inventoryID,
		RouteID:      inventory.RouteID,
		AdjustType:   adjustType,
		OldSpots:     inventory.LeftSpots,
		NewSpots:     newLeftSpots,
		AdjustAmount: adjustAmount,
		Reason:       reason,
		OrderID:      &orderID,
	}

	inventory.LeftSpots = newLeftSpots
	if err := tx.Save(&inventory).Error; err != nil {
		return nil, err
	}

	if err := tx.Create(adjustLog).Error; err != nil {
		return nil, err
	}

	return &inventory, nil
}

func (s *InventoryService) AdjustWithRefund(tx *gorm.DB, inventoryID uint, adjustAmount int, adjustType, reason string, refundID uint, orderID uint) (*models.Inventory, error) {
	var inventory models.Inventory
	if err := tx.First(&inventory, inventoryID).Error; err != nil {
		return nil, err
	}

	newLeftSpots := inventory.LeftSpots + adjustAmount
	if newLeftSpots > inventory.TotalSpots {
		newLeftSpots = inventory.TotalSpots
	}

	adjustLog := &models.InventoryAdjustLog{
		InventoryID:  inventoryID,
		RouteID:      inventory.RouteID,
		AdjustType:   adjustType,
		OldSpots:     inventory.LeftSpots,
		NewSpots:     newLeftSpots,
		AdjustAmount: adjustAmount,
		Reason:       reason,
		RefundID:     &refundID,
		OrderID:      &orderID,
	}

	inventory.LeftSpots = newLeftSpots
	if err := tx.Save(&inventory).Error; err != nil {
		return nil, err
	}

	if err := tx.Create(adjustLog).Error; err != nil {
		return nil, err
	}

	return &inventory, nil
}

func (s *InventoryService) ListAdjustLogs(routeID uint) ([]models.InventoryAdjustLog, error) {
	var logs []models.InventoryAdjustLog
	if err := config.DB.Where("route_id = ?", routeID).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, err
	}
	return logs, nil
}

func (s *InventoryService) UpdateTotalSpots(inventoryID uint, totalSpots int, reason string) (*models.Inventory, error) {
	var inventory models.Inventory
	if err := config.DB.First(&inventory, inventoryID).Error; err != nil {
		return nil, err
	}

	oldTotalSpots := inventory.TotalSpots
	spotDiff := totalSpots - oldTotalSpots

	tx := config.DB.Begin()

	adjustType := "increase_total"
	if spotDiff < 0 {
		adjustType = "decrease_total"
		usedSpots := oldTotalSpots - inventory.LeftSpots
		if totalSpots < usedSpots {
			tx.Rollback()
			return nil, nil
		}
	}

	_, err := s.AdjustWithTx(tx, inventoryID, spotDiff, adjustType, reason)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	inventory.TotalSpots = totalSpots
	if err := tx.Save(&inventory).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return &inventory, nil
}
