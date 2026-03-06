package config

import (
	"log"
	"stajyer-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() {
	//push yapılacağı zaman commentli yer açılacak,açık olan yer silinecek
	dsn := "host=6.tcp.eu.ngrok.io user=postgres password=Canan3905. dbname=stajyer_auth port=10931 sslmode=disable"
	//dsn := "host=192.168.1.23 user=postgres password=Canan3905. dbname=stajyer_auth port=5432 sslmode=disable"

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Veritabanına bağlanılamadı:", err)
	}

	// Tabloları otomatik oluştur (Migration)
	// YENİ EKLENEN: Öğrenci ve Şirket tabloları buraya dahil edildi
	err = DB.AutoMigrate(
		&models.User{},
		&models.Student{}, // EKLENDİ
		&models.Company{}, // EKLENDİ
		&models.OTP{},
		&models.Application{},
		&models.JobPosting{},
		&models.Post{},
		&models.Connection{},
		&models.Message{},
		&models.ContactMessage{},
		&models.DailyReport{},
		&models.Announcement{},
	)

	if err != nil {
		log.Fatal("Migration hatası:", err)
	}

	log.Println("Veritabanı bağlantısı ve migration (Student ve Company dahil) başarılı.")
}
