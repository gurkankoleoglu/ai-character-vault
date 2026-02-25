package main

import (
	"stajyer-backend/config"
	"stajyer-backend/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Veritabanı bağlantısını kur ve tabloları oluştur/güncelle
	config.InitDatabase()

	// 2. Gin framework'ünü başlat
	r := gin.Default()

	// 3. Statik dosya servisini başlat (Kritik: CV ve Mektupların açılması için)
	// Bu satır sayesinde tarayıcıdan /uploads/cv_... linklerine erişilebilir.
	r.Static("/uploads", "./uploads")

	// 4. Rotaları (URL yollarını) tanımla
	routes.SetupRoutes(r)

	// 5. Sunucuyu 8080 portunda çalıştır
	r.Run(":8080")
}
