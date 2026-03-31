package routes

import (
	"stajyer-backend/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(cors.Default())

	//r.Static("/uploads", "./uploads")

	// --- AUTH ---
	r.POST("/login", handlers.Login)
	r.POST("/logout", handlers.Logout)
	r.POST("/request-otp", handlers.RequestOTP)
	r.POST("/verify-otp", handlers.VerifyOTP)
	r.POST("/complete-profile", handlers.CompleteProfile)

	// --- KULLANICI ---
	// ✅ /users/email/:email MUTLAKA /users/:phone'dan ÖNCE gelmeli
	r.GET("/users/email/:email", handlers.GetUserByEmail)
	r.GET("/users/:phone", handlers.GetUserByPhone)
	r.PUT("/users/:phone", handlers.UpdateUserProfile)

	// ✅ Tek tanım — CheckApplicationStatus kaldırıldı
	r.GET("/check-application/:phone", handlers.CheckApplication)

	// --- BAŞVURULAR ---
	r.POST("/apply", handlers.CreateApplication)
	r.GET("/applications", handlers.GetApplications)
	r.GET("/applications/:id", handlers.GetApplicationDetail)
	r.PUT("/applications/:id", handlers.UpdateApplicationStatus)
	r.DELETE("/applications/:id", handlers.DeleteApplication)

	// --- STAJ TAKİP ---
	r.GET("/intern-progress/:phone", handlers.GetInternProgress)

	// --- RAPORLAR ---
	r.POST("/save-report", handlers.SaveDailyReport)
	r.GET("/user-reports/:phone", handlers.GetUserReports)
	r.PUT("/update-report/:id", handlers.UpdateReport)
	r.DELETE("/delete-report/:id", handlers.DeleteReport)

	// --- DUYURULAR ---
	r.POST("/announcements", handlers.CreateAnnouncement)
	r.GET("/announcements", handlers.GetAnnouncements)
	r.GET("/announcements/:id", handlers.GetAnnouncementDetail)
	r.PUT("/announcements/:id", handlers.UpdateAnnouncement)
	r.DELETE("/announcements/:id", handlers.DeleteAnnouncement)

	// --- İLANLAR ---
	r.POST("/jobs", handlers.CreateJob)
	r.GET("/jobs", handlers.GetAllJobs)
	r.PUT("/jobs/:id", handlers.UpdateJob)
	r.DELETE("/jobs/:id", handlers.DeleteJob)

	// --- AKİŞ / SOSYAL ---
	r.GET("/feed", handlers.GetFeed)
	r.POST("/posts/:id/like", handlers.ToggleLike)
	r.POST("/posts/:id/comment", handlers.AddComment)
	r.DELETE("/comments/:id", handlers.DeleteComment)

	// --- BAĞLANTILAR ---
	// ✅ Sabit path'ler (/connect/pending, /connect/list) önce gelmeli
	r.GET("/connect/pending/:phone", handlers.GetPendingRequests)
	r.GET("/connect/list/:phone", handlers.GetConnections)
	r.GET("/connect/status/:phone1/:phone2", handlers.GetConnectionStatus)
	r.POST("/connect", handlers.SendConnectionRequest)
	r.PUT("/connect/:id", handlers.RespondConnectionRequest)
	r.DELETE("/connect/:phone1/:phone2", handlers.RemoveConnection)

	// --- MESAJLAR ---
	// ✅ Sabit path'ler (/messages/inbox, /messages/conversation) önce gelmeli
	r.GET("/messages/inbox/:phone", handlers.GetInbox)
	r.DELETE("/messages/conversation/:phone1/:phone2", handlers.DeleteConversation)
	r.POST("/messages", handlers.SendMessage)
	r.GET("/messages/:phone1/:phone2", handlers.GetConversation)
	r.PUT("/messages/:id", handlers.EditMessage)
	r.DELETE("/messages/:id", handlers.DeleteMessage)

	// --- ADMİN ---
	r.GET("/admin/pending", handlers.GetPendingApprovals)
	r.PUT("/admin/approve-student/:id", handlers.ApproveStudent)
	r.PUT("/admin/approve-company/:id", handlers.ApproveCompany)

	// --- ŞİRKET STAJYER TAKİP ---
	r.GET("/company/:phone/interns", handlers.GetCompanyInterns)
	r.GET("/company/:phone/interns/:intern_user_id/reports", handlers.GetCompanyInternReports)

	// --- EĞİTİM VİDEOLARI ---
	r.GET("/training-videos", handlers.GetPublicVideos)
	r.POST("/training-videos", handlers.CreatePublicVideo)
	r.PUT("/training-videos/:id", handlers.UpdateVideo)
	r.DELETE("/training-videos/:id", handlers.DeleteVideo)

	r.GET("/company-videos", handlers.GetCompanyVideos)
	r.POST("/company-videos", handlers.CreateCompanyVideo)
	r.PUT("/company-videos/:id", handlers.UpdateVideo)
	r.DELETE("/company-videos/:id", handlers.DeleteVideo)

	// --- İLETİŞİM ---
	r.POST("/contact", handlers.SendContactMessage)
}
