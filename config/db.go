package config

import (
	"fmt"
	"os"

	"github.com/Nitesh-04/realtime-racing/models"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"log"
	"time"
)

var DB *gorm.DB

func ConnectDB() {

	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dsn := os.Getenv("DATABASE_URL")

	if dsn == "" {
		log.Fatal("DATABASE_URL is not set in .env file")
	}

	db,err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Room{}, &models.Results{})

	if err != nil {
		log.Fatalf("Error migrating database: %v", err)
	}

	DB = db
	fmt.Println("Database connected successfully")
	fmt.Println("Database Migrated Successfully")

	sqlDB, err := db.DB()

	if err != nil {
		log.Fatalf("Error getting database instance: %v", err)
	}

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

}