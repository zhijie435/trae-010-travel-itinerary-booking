package routes

import (
	"travel-refund/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:5175", "http://localhost:5176", "http://localhost:5177", "http://localhost:5178", "http://localhost:5179", "http://localhost:5180", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	seedCtrl := controllers.NewSeedController()
	userCtrl := controllers.NewUserController()

	r.GET("/api/seed", seedCtrl.SeedData)
	r.GET("/api/users", userCtrl.GetUsers)

	api := r.Group("/api")
	{
		orderCtrl := controllers.NewOrderController()
		orders := api.Group("/orders")
		{
			orders.GET("", orderCtrl.GetOrders)
			orders.GET("/:id", orderCtrl.GetOrder)
			orders.POST("", orderCtrl.CreateOrder)
			orders.POST("/:id/pay", orderCtrl.PayOrder)
		}

		refunds := api.Group("/refunds")
		{
			refunds.GET("", orderCtrl.GetRefundRequests)
			refunds.GET("/:id", orderCtrl.GetRefundRequest)
			refunds.GET("/:id/review-logs", orderCtrl.GetRefundReviewLogs)
			refunds.POST("", orderCtrl.CreateRefundRequest)
			refunds.POST("/:id/review", orderCtrl.ReviewRefundRequest)
			refunds.POST("/batch-review", orderCtrl.BatchReviewRefundRequests)
		}

		routeCtrl := controllers.NewRouteController()
		routes := api.Group("/routes")
		{
			routes.GET("", routeCtrl.GetRoutes)
			routes.GET("/:id", routeCtrl.GetRoute)
			routes.POST("", routeCtrl.CreateRoute)
			routes.PUT("/:id", routeCtrl.UpdateRoute)
			routes.DELETE("/:id", routeCtrl.DeleteRoute)
			routes.GET("/:id/itineraries", routeCtrl.GetItineraries)
			routes.POST("/:id/itineraries", routeCtrl.CreateItinerary)
			routes.PUT("/:id/itineraries/:itinerary_id", routeCtrl.UpdateItinerary)
			routes.DELETE("/:id/itineraries/:itinerary_id", routeCtrl.DeleteItinerary)
		}

		inventoryCtrl := controllers.NewInventoryController()
		inventories := api.Group("/inventories")
		{
			inventories.GET("/route/:id", inventoryCtrl.GetInventories)
			inventories.GET("/:id", inventoryCtrl.GetInventory)
			inventories.POST("/:id/adjust", inventoryCtrl.AdjustSpots)
			inventories.GET("/route/:id/logs", inventoryCtrl.GetAdjustLogs)
		}

		trips := api.Group("/trips")
		{
			trips.GET("", routeCtrl.GetRoutes)
			trips.GET("/:id", routeCtrl.GetRoute)
			trips.POST("", routeCtrl.CreateRoute)
			trips.PUT("/:id", routeCtrl.UpdateRoute)
			trips.DELETE("/:id", routeCtrl.DeleteRoute)
			trips.POST("/:id/adjust-spots", inventoryCtrl.AdjustSpots)
			trips.GET("/:id/itineraries", routeCtrl.GetItineraries)
			trips.POST("/:id/itineraries", routeCtrl.CreateItinerary)
			trips.PUT("/:id/itineraries/:itinerary_id", routeCtrl.UpdateItinerary)
			trips.DELETE("/:id/itineraries/:itinerary_id", routeCtrl.DeleteItinerary)
			trips.GET("/:id/spot-logs", inventoryCtrl.GetAdjustLogs)
		}
	}

	return r
}
