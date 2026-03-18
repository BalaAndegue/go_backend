package config

import (
	"fmt"
	"log"
	"os"

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

	if host == "" { host = "localhost" }
	if user == "" { user = "postgres" }
	if password == "" { password = "postgres" }
	if dbname == "" { dbname = "shopcart_db" }
	if port == "" { port = "5432" }

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port,
	)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Échec de la connexion à la base de données :", err)
	}

	err = database.AutoMigrate(&models.User{}, &models.Category{}, &models.Product{})
	if err != nil {
		log.Fatal("Échec de la migration :", err)
	}

	DB = database
	log.Println("Connexion à la base de données et migrations réussies")
}
