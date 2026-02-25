package routes

import (
	"stajyer-backend/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Frontend'in (port 5173) Backend (port 8080) ile konuşmasına izin verir
	r.Use(cors.Default())

	// --- Auth (Giriş) Rotaları ---
	r.POST("/request-otp", handlers.RequestOTP)
	r.POST("/verify-otp", handlers.VerifyOTP)

	// --- Başvuru Rotaları ---
	// 1. Yeni başvuru oluşturma
	r.POST("/apply", handlers.CreateApplication)

	// 2. Tüm başvuruları listeleme (Admin Panel ana ekranı)
	r.GET("/applications", handlers.GetApplications)

	// 3. Başvuru durumunu güncelleme (Onayla/Reddet butonları için)
	// PATCH kullanılır çünkü sadece 'durum' alanını kısmi olarak güncelliyoruz
	r.PATCH("/applications/:id/status", handlers.UpdateApplicationStatus)

	// 4. Başvuru bilgilerini düzenleme (Düzenle butonu/formu için)
	// PUT kullanılır çünkü tüm başvuru verilerini güncelliyoruz
	r.PUT("/applications/:id", handlers.UpdateApplication)

	// 5. Başvuruyu silme (Sil butonu için)
	r.DELETE("/applications/:id", handlers.DeleteApplication)
}
