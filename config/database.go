package config

import (
	"log"
	"stajyer-backend/models" // Modül adınızın stajyer-backend olduğundan emin olun

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() {
	// Docker üzerindeki PostgreSQL bağlantı bilgileri
	dsn := "host=localhost user=postgres password=Canan3905. dbname=stajyer_auth port=5432 sslmode=disable"

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Veritabanına bağlanılamadı:", err)
	}

	// Tabloları otomatik oluştur (Migration)
	err = DB.AutoMigrate(
		&models.User{},
		&models.OTP{},
		&models.Application{},    // Yeni başvuru tablosu buraya eklendi
		&models.ContactMessage{}, // ContactMessage tablosu eklendi
	)

	if err != nil {
		log.Fatal("Migration hatası:", err)
	}

	log.Println("Veritabanı bağlantısı ve migration başarılı.")
}
