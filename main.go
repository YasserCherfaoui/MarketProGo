package main

import (
	"time"

	"github.com/YasserCherfaoui/MarketProGo/database"
	"github.com/YasserCherfaoui/MarketProGo/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	config := cors.Config{
		AllowOrigins:     []string{"*", "http://localhost:5173", "http://127.0.0.1:5173"}, // Adjust origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // Cache preflight request for 12 hours
	}

	r.Use(cors.New(config))
	db, err := database.ConnectDB()
	if err != nil {
		panic(err)
	}

	routes.AppRoutes(r, db)
	r.Run()
}
