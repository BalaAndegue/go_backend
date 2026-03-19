package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"shopcart-api/config"
	"shopcart-api/routes"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Pas de fichier .env trouvé")
	}

	config.ConnectDatabase()

	r := gin.Default()

	// CORS Configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Serve static files (images)
	r.Static("/uploads", "./uploads")

	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	r.Run(":" + port)
}
