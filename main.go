package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	// CORS Configuration.
	// A wildcard origin cannot be combined with credentials (browsers reject it
	// and gin-contrib/cors errors out). So credentials are only enabled when an
	// explicit allow-list is provided via CORS_ALLOWED_ORIGINS (comma-separated).
	corsConfig := cors.Config{
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders: []string{"Content-Length"},
	}
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		corsConfig.AllowOrigins = strings.Split(origins, ",")
		corsConfig.AllowCredentials = true
	} else {
		corsConfig.AllowOrigins = []string{"*"}
	}
	r.Use(cors.New(corsConfig))

	// Serve static files (images)
	r.Static("/uploads", "./uploads")

	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Run the server in a goroutine so it does not block the shutdown handling.
	go func() {
		log.Printf("Serveur démarré sur le port %s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Erreur serveur : %v", err)
		}
	}()

	// Wait for an interrupt/termination signal, then shut down gracefully.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Arrêt du serveur en cours...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Arrêt forcé du serveur : %v", err)
	}
	log.Println("Serveur arrêté proprement")
}
