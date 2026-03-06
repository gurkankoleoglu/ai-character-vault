package handlers

import (
	"fmt"
	"net/http"
	"stajyer-backend/config"
	"stajyer-backend/models"
	"stajyer-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// RequestOTP: Kullanıcı telefon girince durumu kontrol eder, onaylı değilse OTP göndermez
func RequestOTP(c *gin.Context) {
	type Request struct {
		Phone string `json:"phone" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
		return
	}

	if !utils.IsValidTurkishPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefon 10 hane olmalı ve 5 ile başlamalıdır"})
		return
	}

	// 1. ADMIN VE BAŞVURU KONTROLÜ
	var user models.User
	isAdmin := false
	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err == nil && user.Role == "admin" {
		isAdmin = true
	}

	if !isAdmin {
		var app models.Application
		if err := config.DB.Where("phone = ?", req.Phone).Order("created_at desc").First(&app).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Henüz bir başvurunuz bulunmamaktadır. Lütfen önce başvuru yapınız."})
			return
		}

		// ⭐ STATICCHECK UYARISINI GİDEREN SWITCH YAPISI (QF1003)
		switch app.Durum {
		case "Beklemede":
			c.JSON(http.StatusForbidden, gin.H{"error": "Başvurunuz henüz değerlendirme aşamasındadır. Onaylandıktan sonra giriş yapabilirsiniz."})
			return
		case "Reddedildi":
			c.JSON(http.StatusForbidden, gin.H{"error": "Başvurunuz reddedilmiştir. Sisteme giriş yetkiniz bulunmamaktadır."})
			return
		}
	}

	// 2. SPAM KONTROLÜ
	var lastOTP models.OTP
	err := config.DB.Where("phone = ?", req.Phone).Order("created_at DESC").First(&lastOTP).Error
	if err == nil && time.Since(lastOTP.CreatedAt) < 60*time.Second {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Lütfen yeni kod için 60 saniye bekleyin."})
		return
	}

	// 3. OTP ÜRETME
	otpCode := utils.GenerateOTP()
	fmt.Printf("\n🚀 [SMS SİMÜLASYONU] -> Tel: %s | Kod: %s\n", req.Phone, otpCode)

	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)
	otp := models.OTP{
		Phone:     req.Phone,
		Code:      string(hashedCode),
		ExpiresAt: time.Now().Add(2 * time.Minute),
		IsUsed:    false,
	}

	config.DB.Create(&otp)
	c.JSON(http.StatusOK, gin.H{"message": "Doğrulama kodu oluşturuldu.", "otp": otpCode})
}

// VerifyOTP: Kod doğrulama sonrası kullanıcıyı sisteme alır
func VerifyOTP(c *gin.Context) {
	type Request struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri formatı"})
		return
	}

	var otp models.OTP
	err := config.DB.Where("phone = ? AND is_used = ?", req.Phone, false).Order("created_at DESC").First(&otp).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçerli bir kod bulunamadı"})
		return
	}

	if time.Now().After(otp.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kodun süresi dolmuş"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(otp.Code), []byte(req.Code)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Hatalı kod girdiniz"})
		return
	}

	otp.IsUsed = true
	config.DB.Save(&otp)

	var user models.User
	isNewUser := false

	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		user = models.User{Phone: req.Phone, Role: "user"}
		config.DB.Create(&user)
		isNewUser = true
	} else if user.FullNames == "" {
		isNewUser = true
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Giriş başarılı!",
		"user_id":     user.ID,
		"phone":       user.Phone,
		"role":        user.Role,
		"is_new_user": isNewUser,
	})
}

// CompleteProfile: Profil bilgilerini günceller
func CompleteProfile(c *gin.Context) {
	type ProfileUpdateRequest struct {
		Phone      string `json:"phone" binding:"required"`
		TCKimlik   string `json:"tc_kimlik"`
		FullName   string `json:"full_name"`
		Email      string `json:"email"`
		University string `json:"university"`
		Department string `json:"department"`
		InternType string `json:"intern_type"`
		Gender     string `json:"gender"`
	}

	var req ProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Veri formatı hatalı"})
		return
	}

	var user models.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	user.FullNames = req.FullName
	user.Email = req.Email
	user.TCNo = req.TCKimlik
	user.University = req.University
	user.Department = req.Department
	user.IsProfileCompleted = true

	config.DB.Save(&user)
	c.JSON(http.StatusOK, gin.H{"message": "Profil güncellendi", "user": user})
}

// GetUserByPhone: Tekil kullanıcı bilgisini getirir
func GetUserByPhone(c *gin.Context) {
	phone := c.Param("phone")
	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}
	c.JSON(http.StatusOK, user)
}
