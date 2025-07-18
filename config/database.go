package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/BilyHakim/go-walet/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB() *gorm.DB {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s Timezone=Asia/Jakarta",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	log.Printf("Attempting to connect to database with DSN: %s", dsn)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	log.Printf("Database connection established successfully")

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get database connection: %v", err)
	}

	// Test database connection
	err = sqlDB.Ping()
	if err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Printf("Database ping successful")

	log.Printf("Starting database migration...")
	err = db.AutoMigrate(&models.User{}, &models.Transaction{})
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
	log.Printf("Database migrated successfully")

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}
