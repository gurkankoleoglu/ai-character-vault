package config

import (
	"log"
	"stajyer-backend/models" // Proje adınıza göre güncelleyin

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() {
	dsn := "host=localhost user=postgres password=Canan3905. dbname=stajyer_auth port=5432 sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Veritabanına bağlanılamadı:", err)
	}

	DB.AutoMigrate(&models.User{}, &models.OTP{})
}
