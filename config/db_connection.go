package config

import (
	"log"

	"github.com/ronak-89/simple-user-auth/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var cnf, _ = LoadConfig()

var Db *gorm.DB

func DbConnection() *gorm.DB {

	dsn := cnf.DB.URL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	Db = db
	return Db

}

func AutoMigrate() {
	err := Db.AutoMigrate(&models.User{}, &models.EmailOtp{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
}
