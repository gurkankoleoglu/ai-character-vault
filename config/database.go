package config

import (
	"fmt"
	"log"
	"os"
	"stajyer-backend/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() {
	// 1. .env dosyasını yükle. Eğer dosya yoksa hata vermez ama uyarı logu basar.
	// Production ortamlarında genellikle sistem değişkenleri kullanılır.
	err := godotenv.Load()
	if err != nil {
		log.Println(".env dosyası yüklenemedi (Sistem çevresel değişkenleri kullanılıyor olabilir)")
	}

	// 2. Değişkenleri .env dosyasından çek
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")

	// 3. Bağlantı dizesini dinamik oluştur
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, user, password, dbname, port, sslmode)

	// 4. Veritabanı bağlantısını başlat
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Veritabanına bağlanılamadı:", err)
	}

	// Tabloları otomatik oluştur (Migration)
	err = DB.AutoMigrate(
		&models.User{},
		&models.Student{},
		&models.Company{},
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

	log.Println("Veritabanı bağlantısı ve migration (Tüm tablolar dahil) başarılı.")
}
