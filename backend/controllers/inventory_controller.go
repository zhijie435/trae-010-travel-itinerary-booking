package controllers

import (
	"net/http"
	"strconv"
	"travel-refund/services"

	"github.com/gin-gonic/gin"
)

type InventoryController struct {
	inventoryService *services.InventoryService
}

func NewInventoryController() *InventoryController {
	return &InventoryController{
		inventoryService: services.NewInventoryService(),
	}
}

type SpotAdjustInput struct {
	AdjustAmount int    `json:"adjust_amount" binding:"required"`
	Reason       string `json:"reason" binding:"required"`
}

func (ctrl *InventoryController) GetInventories(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	inventories, err := ctrl.inventoryService.ListByRoute(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": inventories})
}

func (ctrl *InventoryController) GetInventory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	inventory, err := ctrl.inventoryService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "库存不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": inventory})
}

func (ctrl *InventoryController) AdjustSpots(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var input SpotAdjustInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.AdjustAmount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "调整数量不能为0"})
		return
	}

	adjustType := "manual_increase"
	if input.AdjustAmount < 0 {
		adjustType = "manual_decrease"
	}

	inventory, err := ctrl.inventoryService.Adjust(uint(id), input.AdjustAmount, adjustType, input.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if inventory == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "调整后余位不能为负数或超过总名额"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": inventory})
}

func (ctrl *InventoryController) GetAdjustLogs(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	logs, err := ctrl.inventoryService.ListAdjustLogs(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": logs})
}
