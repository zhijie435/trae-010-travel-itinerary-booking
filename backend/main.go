package main

import (
	"log"
	"travel-refund/config"
	"travel-refund/routes"
)

func main() {
	config.InitDB()

	r := routes.SetupRouter()

	log.Println("Server starting on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
