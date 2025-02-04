package main

import (
	"github.com/ronak-89/simple-user-auth/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var cnf, _ = config.LoadConfig()

func DbConnection() *gorm.DB {
	dsn := cnf.DB.URL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	return db
}
