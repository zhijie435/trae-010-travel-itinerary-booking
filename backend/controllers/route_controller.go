package controllers

import (
	"net/http"
	"strconv"
	"travel-refund/pkg/utils"
	"travel-refund/services"

	"github.com/gin-gonic/gin"
)

type RouteController struct {
	routeService *services.RouteService
}

func NewRouteController() *RouteController {
	return &RouteController{
		routeService: services.NewRouteService(),
	}
}

type RouteInput struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Destination string  `json:"destination" binding:"required"`
	StartDate   string  `json:"start_date" binding:"required"`
	EndDate     string  `json:"end_date" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
	TotalSpots  int     `json:"total_spots" binding:"required"`
	Status      string  `json:"status"`
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

func (ctrl *RouteController) GetRoutes(c *gin.Context) {
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	routes, err := ctrl.routeService.List(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": routes})
}

func (ctrl *RouteController) GetRoute(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	route, err := ctrl.routeService.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "线路不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": route})
}

func (ctrl *RouteController) CreateRoute(c *gin.Context) {
	var input RouteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := utils.ParseDate(input.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式无效"})
		return
	}
	endDate, err := utils.ParseDate(input.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式无效"})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期不能早于开始日期"})
		return
	}

	route, err := ctrl.routeService.Create(
		input.Name, input.Description, input.Destination,
		startDate, endDate, input.Price, input.TotalSpots, input.Status,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": route})
}

func (ctrl *RouteController) UpdateRoute(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var input RouteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := utils.ParseDate(input.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "开始日期格式无效"})
		return
	}
	endDate, err := utils.ParseDate(input.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期格式无效"})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "结束日期不能早于开始日期"})
		return
	}

	route, err := ctrl.routeService.Update(
		uint(id), input.Name, input.Description, input.Destination,
		startDate, endDate, input.Price, input.TotalSpots, input.Status,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if route == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "总名额不能少于已售出名额"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": route})
}

func (ctrl *RouteController) DeleteRoute(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	err := ctrl.routeService.Delete(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "线路不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

func (ctrl *RouteController) GetItineraries(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	itineraries, err := ctrl.routeService.ListItineraries(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": itineraries})
}

func (ctrl *RouteController) CreateItinerary(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)

	var input ItineraryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svcInput := &services.ItineraryInput{
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

	itinerary, err := ctrl.routeService.CreateItinerary(uint(id), svcInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if itinerary == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该天行程已存在或天数无效"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": itinerary})
}

func (ctrl *RouteController) UpdateItinerary(c *gin.Context) {
	itineraryID, _ := strconv.ParseUint(c.Param("itinerary_id"), 10, 32)

	var input ItineraryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	svcInput := &services.ItineraryInput{
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

	itinerary, err := ctrl.routeService.UpdateItinerary(uint(itineraryID), svcInput)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "行程安排不存在"})
		return
	}
	if itinerary == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该天行程已存在或天数无效"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": itinerary})
}

func (ctrl *RouteController) DeleteItinerary(c *gin.Context) {
	itineraryID, _ := strconv.ParseUint(c.Param("itinerary_id"), 10, 32)

	err := ctrl.routeService.DeleteItinerary(uint(itineraryID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "行程安排不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}
