package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"shopcart-api/models"
)

var DB *gorm.DB

func ConnectDatabase() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Aucun fichier .env trouvé, utilisation des variables d'environnement")
	}

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	if host == "" {
		host = "localhost"
	}
	if user == "" {
		user = "postgres"
	}
	if password == "" {
		password = "postgres"
	}
	if dbname == "" {
		dbname = "shopcart_db"
	}
	if port == "" {
		port = "5432"
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode,
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Échec de la connexion à la base de données :", err)
	}

	// Configure the underlying connection pool with sane production defaults.
	if sqlDB, err := database.DB(); err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(time.Hour)
		sqlDB.SetConnMaxIdleTime(10 * time.Minute)
	}

	// AutoMigrate is convenient in development but is not a safe production
	// migration strategy, so it can be disabled via AUTO_MIGRATE=false.
	if autoMigrateEnabled() {
		err = database.AutoMigrate(
			&models.User{},
			&models.Category{},
			&models.Product{},
			&models.ProductVariant{},
			&models.Cart{},
			&models.CartItem{},
			&models.Order{},
			&models.OrderItem{},
			&models.Payment{},
			&models.DeliveryGeolocation{},
		)
		if err != nil {
			log.Fatal("Échec de la migration :", err)
		}
	}

	DB = database
	log.Println("Connexion à la base de données et migrations réussies")
}

// autoMigrateEnabled reports whether AutoMigrate should run. It defaults to true
// and is disabled only when AUTO_MIGRATE is explicitly set to a false-y value.
func autoMigrateEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("AUTO_MIGRATE"))) {
	case "false", "0", "no", "off":
		return false
	default:
		return true
	}
}
