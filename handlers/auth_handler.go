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

// RequestOTP: Kullanıcı telefon girince OTP üretir ve DB'ye kaydeder
func RequestOTP(c *gin.Context) {
	type Request struct {
		Phone string `json:"phone" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
		return
	}

	// Telefon format kontrolü
	if !utils.IsValidTurkishPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefon 10 hane olmalı ve 5 ile başlamalıdır"})
		return
	}

	// Rate Limit: 60 saniye kuralı
	var lastOTP models.OTP
	err := config.DB.Where("phone = ?", req.Phone).Order("created_at DESC").First(&lastOTP).Error
	if err == nil && time.Since(lastOTP.CreatedAt) < 60*time.Second {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Lütfen yeni kod için 60 saniye bekleyin."})
		return
	}

	otpCode := utils.GenerateOTP()
	fmt.Printf("\n🚀 YENİ OTP -> Tel: %s | Kod: %s\n", req.Phone, otpCode)

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

// VerifyOTP: Gelen kodu doğrular ve kullanıcının profil durumunu döner
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçerli kod bulunamadı"})
		return
	}

	if time.Now().After(otp.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kodun süresi dolmuş"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(otp.Code), []byte(req.Code)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Hatalı kod"})
		return
	}

	// Kod doğrulandı, kullanıldı olarak işaretle
	otp.IsUsed = true
	config.DB.Save(&otp)

	// Kullanıcıyı kontrol et
	var user models.User
	isNewUser := false

	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		// Kullanıcı yoksa oluştur (Auto-Register)
		user = models.User{Phone: req.Phone, Role: "user"}
		config.DB.Create(&user)
		isNewUser = true
	} else {
		// Kullanıcı var ama FullNames (Ad Soyad) alanı boşsa "Yeni Kullanıcı" sayılır
		if user.FullNames == "" {
			isNewUser = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Giriş başarılı!",
		"user_id":     user.ID,
		"phone":       user.Phone,
		"role":        user.Role,
		"is_new_user": isNewUser, // Frontend bu bayrağa göre yönlendirme yapacak
	})
}

// CompleteProfile: Eksik olan profil bilgilerini veritabanına yazar
func CompleteProfile(c *gin.Context) {
	type ProfileUpdateRequest struct {
		Phone      string `json:"phone" binding:"required"`
		TCKimlik   string `json:"tc_kimlik"`
		FullName   string `json:"full_name"`
		Email      string `json:"email"`
		University string `json:"university"`
		Department string `json:"department"`
	}

	var req ProfileUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri formatı"})
		return
	}

	var user models.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı. Lütfen önce giriş yapın."})
		return
	}

	// Verileri Model'deki alanlarla eşle
	user.FullNames = req.FullName
	user.Email = req.Email
	user.TCNo = req.TCKimlik
	user.University = req.University
	user.Department = req.Department
	user.IsProfileCompleted = true

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Profil güncellenirken hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profil başarıyla tamamlandı",
		"user":    user,
	})
}

// GetUserByPhone: Telefon numarasına göre kullanıcı bilgilerini döner (Giriş kontrolü için)
func GetUserByPhone(c *gin.Context) {
	phone := c.Param("phone")
	var user models.User

	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, user)
}
