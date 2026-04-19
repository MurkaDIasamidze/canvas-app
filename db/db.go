package db

import (
	"fmt"
	"log"
	"os"

	"canvas-tui/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Connect() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		env("DB_HOST", "localhost"),
		env("DB_USER", "postgres"),
		env("DB_PASSWORD", "root"),
		env("DB_NAME", "postgres"),
		env("DB_PORT", "5432"),
	)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("DB: %v", err)
	}
	DB.AutoMigrate(&models.Project{}, &models.Shape{})
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}