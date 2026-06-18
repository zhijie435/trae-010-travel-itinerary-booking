package services

import (
	"travel-refund/config"
	"travel-refund/models"
	"travel-refund/pkg/utils"
	"time"
)

type RouteService struct{}

func NewRouteService() *RouteService {
	return &RouteService{}
}

func (s *RouteService) List(startDate, endDate string) ([]models.Route, error) {
	var routes []models.Route
	query := config.DB.Order("created_at DESC")

	if startDate != "" {
		if t, err := utils.ParseDate(startDate); err == nil {
			query = query.Where("julianday(end_date) >= julianday(?)", t.Format(time.RFC3339))
		}
	}
	if endDate != "" {
		if t, err := utils.ParseDate(endDate); err == nil {
			endOfDay := t.Add(24*time.Hour - time.Second)
			query = query.Where("julianday(start_date) <= julianday(?)", endOfDay.Format(time.RFC3339))
		}
	}

	if err := query.Find(&routes).Error; err != nil {
		return nil, err
	}
	return routes, nil
}

func (s *RouteService) GetByID(id uint) (*models.Route, error) {
	var route models.Route
	if err := config.DB.Preload("Itineraries").First(&route, id).Error; err != nil {
		return nil, err
	}
	return &route, nil
}

func (s *RouteService) Create(name, description, destination string, startDate, endDate time.Time, price float64, totalSpots int, status string) (*models.Route, error) {
	route := &models.Route{
		Name:        name,
		Description: description,
		Destination: destination,
		StartDate:   startDate,
		EndDate:     endDate,
		Price:       price,
		Status:      "active",
	}
	if status != "" {
		route.Status = status
	}

	tx := config.DB.Begin()
	if err := tx.Create(route).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	inventoryService := NewInventoryService()
	_, err := inventoryService.CreateWithTx(tx, route.ID, route.StartDate.Format("2006-01-02"), totalSpots, "init", "线路创建初始化")
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return route, nil
}

func (s *RouteService) Update(id uint, name, description, destination string, startDate, endDate time.Time, price float64, totalSpots int, status string) (*models.Route, error) {
	var route models.Route
	if err := config.DB.First(&route, id).Error; err != nil {
		return nil, err
	}

	oldTotalSpots := 0
	var inventory models.Inventory
	if err := config.DB.Where("route_id = ? AND date = ?", id, route.StartDate.Format("2006-01-02")).First(&inventory).Error; err == nil {
		oldTotalSpots = inventory.TotalSpots
	}

	spotDiff := totalSpots - oldTotalSpots

	tx := config.DB.Begin()

	route.Name = name
	route.Description = description
	route.Destination = destination
	route.StartDate = startDate
	route.EndDate = endDate
	route.Price = price

	if spotDiff != 0 && oldTotalSpots > 0 {
		inventoryService := NewInventoryService()
		adjustType := "increase_total"
		reason := "调整总名额增加"
		if spotDiff < 0 {
			adjustType = "decrease_total"
			reason = "调整总名额减少"
			usedSpots := oldTotalSpots - inventory.LeftSpots
			if totalSpots < usedSpots {
				tx.Rollback()
				return nil, nil
			}
		}
		_, err := inventoryService.AdjustWithTx(tx, inventory.ID, spotDiff, adjustType, reason)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if status != "" {
		route.Status = status
	}

	if err := tx.Save(&route).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return &route, nil
}

func (s *RouteService) Delete(id uint) error {
	var route models.Route
	if err := config.DB.First(&route, id).Error; err != nil {
		return err
	}

	var orderCount int64
	config.DB.Model(&models.Order{}).Where("route_id = ?", id).Count(&orderCount)
	if orderCount > 0 {
		return nil
	}

	tx := config.DB.Begin()
	if err := tx.Where("route_id = ?", id).Delete(&models.Itinerary{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("route_id = ?", id).Delete(&models.InventoryAdjustLog{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Where("route_id = ?", id).Delete(&models.Inventory{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Delete(&route).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return nil
}

func (s *RouteService) ListItineraries(routeID uint) ([]models.Itinerary, error) {
	var itineraries []models.Itinerary
	if err := config.DB.Where("route_id = ?", routeID).Order("day_number ASC").Find(&itineraries).Error; err != nil {
		return nil, err
	}
	return itineraries, nil
}

func (s *RouteService) CreateItinerary(routeID uint, input *ItineraryInput) (*models.Itinerary, error) {
	var route models.Route
	if err := config.DB.First(&route, routeID).Error; err != nil {
		return nil, err
	}

	if input.DayNumber < 1 {
		return nil, nil
	}

	var existing models.Itinerary
	result := config.DB.Where("route_id = ? AND day_number = ?", routeID, input.DayNumber).First(&existing)
	if result.Error == nil {
		return nil, nil
	}

	itinerary := &models.Itinerary{
		RouteID:        routeID,
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

	if err := config.DB.Create(itinerary).Error; err != nil {
		return nil, err
	}

	return itinerary, nil
}

func (s *RouteService) UpdateItinerary(itineraryID uint, input *ItineraryInput) (*models.Itinerary, error) {
	var itinerary models.Itinerary
	if err := config.DB.First(&itinerary, itineraryID).Error; err != nil {
		return nil, err
	}

	if input.DayNumber < 1 {
		return nil, nil
	}

	if input.DayNumber != itinerary.DayNumber {
		var existing models.Itinerary
		result := config.DB.Where("route_id = ? AND day_number = ? AND id != ?", itinerary.RouteID, input.DayNumber, itineraryID).First(&existing)
		if result.Error == nil {
			return nil, nil
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
		return nil, err
	}

	return &itinerary, nil
}

func (s *RouteService) DeleteItinerary(itineraryID uint) error {
	var itinerary models.Itinerary
	if err := config.DB.First(&itinerary, itineraryID).Error; err != nil {
		return err
	}

	return config.DB.Delete(&itinerary).Error
}

type ItineraryInput struct {
	DayNumber      int
	Title          string
	Breakfast      string
	Lunch          string
	Dinner         string
	Accommodation  string
	Transportation string
	Activities     string
	Notes          string
}
