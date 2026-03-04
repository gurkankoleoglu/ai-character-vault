package routes

import (
	"stajyer-backend/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Frontend (5173) ve Backend (8080) arasındaki iletişimi sağlar
	r.Use(cors.Default())

	// --- 1. KULLANICI GİRİŞ (AUTH) ROTALARI ---
	r.POST("/request-otp", handlers.RequestOTP)
	r.POST("/verify-otp", handlers.VerifyOTP)

	// --- 2. STAJ BAŞVURU ROTALARI ---
	// Yeni staj başvurusu al
	r.POST("/apply", handlers.CreateApplication)

	// Tüm başvuruları listele (Admin Panel)
	r.GET("/applications", handlers.GetApplications)

	// Başvuru durumunu (Onayla/Reddet) güncelle
	r.PATCH("/applications/:id/status", handlers.UpdateApplicationStatus)

	// Başvuru bilgilerini düzenle
	// r.PUT("/applications/:id", handlers.UpdateApplication)

	// Başvuruyu sistemden sil
	r.DELETE("/applications/:id", handlers.DeleteApplication)

	// --- 3. İLETİŞİM FORMU ROTASI ---
	// Ana sayfadaki iletişim formundan gelen mesajları kaydeder
	// Terminaldeki 404 hatasını çözen satır budur
	r.POST("/contact", handlers.SendContactMessage)

	// SetupRoutes içine ekle:
	r.POST("/complete-profile", handlers.CompleteProfile)
	// SetupRoutes içine ekle:
	r.GET("/applications/:id", handlers.GetApplicationDetail)
	// routes.go içine ekle
	r.GET("/users/:phone", handlers.GetUserByPhone)
	// Mevcut rotalarının arasına ekle
	r.GET("/intern-progress/:phone", handlers.GetInternProgress)
	// Rapor kaydetmek için POST rotası
	r.POST("/save-report", handlers.SaveDailyReport)
	r.GET("/user-reports/:phone", handlers.GetUserReports)
	// Mevcut rotalarının yanına ekle
	r.PUT("/update-report/:id", handlers.UpdateReport)
	r.DELETE("/delete-report/:id", handlers.DeleteReport)
}
