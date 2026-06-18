package routes

import (
	"travel-refund/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	r.GET("/api/seed", controllers.SeedData)
	r.GET("/api/trips", controllers.GetTrips)
	r.GET("/api/users", controllers.GetUsers)

	api := r.Group("/api")
	{
		orders := api.Group("/orders")
		{
			orders.GET("", controllers.GetOrders)
			orders.GET("/:id", controllers.GetOrder)
			orders.POST("", controllers.CreateOrder)
			orders.POST("/:id/pay", controllers.PayOrder)
		}

		refunds := api.Group("/refunds")
		{
			refunds.GET("", controllers.GetRefundRequests)
			refunds.GET("/:id", controllers.GetRefundRequest)
			refunds.POST("", controllers.CreateRefundRequest)
			refunds.POST("/:id/review", controllers.ReviewRefundRequest)
		}
	}

	return r
}
