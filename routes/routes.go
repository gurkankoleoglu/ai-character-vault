package routes

import (
	"stajyer-backend/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Bu satır frontend'in backend ile konuşmasını sağlar
	r.Use(cors.Default())

	r.POST("/request-otp", handlers.RequestOTP)
	r.POST("/verify-otp", handlers.VerifyOTP)
	// Başvuru Rotası (Yeni)
	r.POST("/apply", handlers.CreateApplication)
}
