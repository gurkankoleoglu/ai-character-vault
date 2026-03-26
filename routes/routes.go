/*
*package routes

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
		r.POST("/contact", handlers.SendContactMessage)

		// SetupRoutes içine ekle:
		r.POST("/complete-profile", handlers.CompleteProfile)
		r.GET("/applications/:id", handlers.GetApplicationDetail)
		r.GET("/users/:phone", handlers.GetUserByPhone)
		// Mevcut rotalarının arasına ekle
		r.GET("/intern-progress/:phone", handlers.GetInternProgress)
		// Rapor kaydetmek için POST rotası
		r.POST("/save-report", handlers.SaveDailyReport)
		r.GET("/user-reports/:phone", handlers.GetUserReports)
		// Mevcut rotalarının yanına ekle
		r.PUT("/update-report/:id", handlers.UpdateReport)
		r.DELETE("/delete-report/:id", handlers.DeleteReport)

		// --- 4. DUYURU ROTALARI ---
		// Yeni bir duyuru oluştur
		r.POST("/announcements", handlers.CreateAnnouncement)

		// Tüm duyuruları listele
		r.GET("/announcements", handlers.GetAnnouncements)

		// Belirli bir duyuruyu getir
		r.GET("/announcements/:id", handlers.GetAnnouncementDetail)

		// Duyuruyu güncelle
		r.PUT("/announcements/:id", handlers.UpdateAnnouncement)

		// Duyuruyu sil
		r.DELETE("/announcements/:id", handlers.DeleteAnnouncement)
	}

*
*/
package routes

import (
	"stajyer-backend/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Frontend (5173) ve Backend (8080) arasındaki iletişimi sağlar
	r.Use(cors.Default())

	// --- STATİK DOSYA SUNUMU (YENİ EKLENDİ) ---
	// Bu satır, yüklenen resimlerin (veya CV'lerin) frontend tarafında
	// "http://localhost:8080/uploads/..." şeklinde görüntülenmesini sağlar.

	//r.Static("/uploads", "./uploads")

	// --- 1. KULLANICI GİRİŞ (AUTH) ROTALARI ---
	r.POST("/request-otp", handlers.RequestOTP)
	r.POST("/verify-otp", handlers.VerifyOTP)

	// --- 2. STAJ BAŞVURU ROTALARI ---
	// Yeni staj başvurusu al
	r.POST("/apply", handlers.CreateApplication)

	// Tüm başvuruları listele (Admin Panel)
	r.GET("/applications", handlers.GetApplications)

	// Başvuru durumunu (Onayla/Reddet) güncelle
	r.PUT("/applications/:id", handlers.UpdateApplicationStatus)

	// Başvuru bilgilerini düzenle
	// r.PUT("/applications/:id", handlers.UpdateApplication)

	// Başvuruyu sistemden sil
	r.DELETE("/applications/:id", handlers.DeleteApplication)

	// Kullanıcının başvuru durumunu kontrol et
	r.GET("/check-application/:phone", handlers.CheckApplicationStatus)

	// --- 3. İLETİŞİM FORMU ROTASI ---
	// Ana sayfadaki iletişim formundan gelen mesajları kaydeder
	r.POST("/contact", handlers.SendContactMessage)

	// SetupRoutes içine ekle:
	r.POST("/complete-profile", handlers.CompleteProfile)
	r.GET("/applications/:id", handlers.GetApplicationDetail)
	r.GET("/users/:phone", handlers.GetUserByPhone)

	// Mevcut rotalarının arasına ekle
	r.GET("/intern-progress/:phone", handlers.GetInternProgress)

	// Rapor kaydetmek için POST rotası
	r.POST("/save-report", handlers.SaveDailyReport)
	r.GET("/user-reports/:phone", handlers.GetUserReports)

	// Mevcut rotalarının yanına ekle
	r.PUT("/update-report/:id", handlers.UpdateReport)
	r.DELETE("/delete-report/:id", handlers.DeleteReport)

	// --- 4. DUYURU ROTALARI ---
	// Yeni bir duyuru oluştur
	r.POST("/announcements", handlers.CreateAnnouncement)

	// Tüm duyuruları listele
	r.GET("/announcements", handlers.GetAnnouncements)

	// Belirli bir duyuruyu getir
	r.GET("/announcements/:id", handlers.GetAnnouncementDetail)

	// Duyuruyu güncelle
	r.PUT("/announcements/:id", handlers.UpdateAnnouncement)

	// Duyuruyu sil
	r.DELETE("/announcements/:id", handlers.DeleteAnnouncement)

	r.POST("/jobs", handlers.CreateJob)
	r.GET("/jobs", handlers.GetAllJobs)

	// YENİ EKLENEN İLAN DÜZENLEME VE SİLME ROTALARI:
	r.PUT("/jobs/:id", handlers.UpdateJob)
	r.DELETE("/jobs/:id", handlers.DeleteJob)

	r.POST("/posts", handlers.CreatePost)
	r.GET("/feed", handlers.GetFeed)

	// YENİ EKLENEN SOSYAL AĞ ETKİLEŞİM ROTALARI
	r.POST("/posts/:id/like", handlers.ToggleLike)
	r.POST("/posts/:id/comment", handlers.AddComment)
	// Belirli bir kullanıcının gönderilerini çek (Profil sayfası için)
	r.GET("/posts/user/:phone", handlers.GetUserPosts)

	r.POST("/connect", handlers.SendConnectionRequest)            // İstek At
	r.PUT("/connect/:id", handlers.RespondConnectionRequest)      // Kabul Et/Reddet
	r.GET("/connect/pending/:phone", handlers.GetPendingRequests) // Gelen İsteklerimi Gör

	r.POST("/messages", handlers.SendMessage)                    // Mesaj Gönder
	r.GET("/messages/:phone1/:phone2", handlers.GetConversation) // Sohbet Geçmişini Çek

	// YENİ EKLENEN İNBOX ROTASI
	r.GET("/messages/inbox/:phone", handlers.GetInbox)

	// YENİ EKLENEN MESAJ İŞLEMLERİ (Düzenle, Sil, Sohbeti Sil)
	r.PUT("/messages/:id", handlers.EditMessage)                                    // Mesajı Düzenle
	r.DELETE("/messages/:id", handlers.DeleteMessage)                               // Tekil Mesaj Sil
	r.DELETE("/messages/conversation/:phone1/:phone2", handlers.DeleteConversation) // Tüm Sohbeti Sil

	// Admin İşlemleri
	r.GET("/admin/pending", handlers.GetPendingApprovals)
	r.PUT("/admin/approve-student/:id", handlers.ApproveStudent)
	r.PUT("/admin/approve-company/:id", handlers.ApproveCompany)

	// Şirket - Stajyer Raporları İzleme
	r.GET("/company/:phone/interns", handlers.GetCompanyInterns)
	r.GET("/company/:phone/interns/:intern_user_id/reports", handlers.GetCompanyInternReports)

	// Profil Düzenleme İşlemi
	r.PUT("/users/:phone", handlers.UpdateUserProfile)

	// Bağlantılar
	r.GET("/connect/list/:phone", handlers.GetConnections) // Onaylı bağlantılarım

	r.GET("/connect/status/:phone1/:phone2", handlers.GetConnectionStatus) // İki kullanıcı arasındaki bağlantı durumunu kontrol et

	r.POST("/logout", handlers.Logout)

	// ─── EĞİTİM VİDEOLARI ROTALARI ───
	r.GET("/training-videos", handlers.GetPublicVideos)
	r.POST("/training-videos", handlers.CreatePublicVideo)
	r.PUT("/training-videos/:id", handlers.UpdateVideo)
	r.DELETE("/training-videos/:id", handlers.DeleteVideo)

	r.GET("/company-videos", handlers.GetCompanyVideos)
	r.POST("/company-videos", handlers.CreateCompanyVideo)
	r.PUT("/company-videos/:id", handlers.UpdateVideo)
	r.DELETE("/company-videos/:id", handlers.DeleteVideo)

}
