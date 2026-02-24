package handlers

import (
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
		Phone string `json:"phone"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
		return
	}

	// Telefon format kontrolü (Utils'den çekiyoruz)
	if !utils.IsValidTurkishPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefon 10 hane olmalı ve 5 ile başlamalıdır"})
		return
	}

	// Rate Limit: 60 saniyede bir OTP
	var lastOTP models.OTP
	err := config.DB.Where("phone = ?", req.Phone).Order("created_at DESC").First(&lastOTP).Error
	if err == nil && time.Since(lastOTP.CreatedAt) < 60*time.Second {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Lütfen 60 saniye bekleyin."})
		return
	}

	otpCode := utils.GenerateOTP()
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)

	otp := models.OTP{
		Phone:     req.Phone,
		Code:      string(hashedCode),
		ExpiresAt: time.Now().Add(2 * time.Minute),
	}

	if err := config.DB.Create(&otp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OTP kaydedilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP oluşturuldu", "otp": otpCode})
}

// VerifyOTP: Gelen kodu doğrular ve kullanıcıyı giriş yaptırır/kaydeder
func VerifyOTP(c *gin.Context) {
	type Request struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
		return
	}

	var otp models.OTP
	err := config.DB.Where("phone = ? AND is_used = ?", req.Phone, false).Order("created_at DESC").First(&otp).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kod bulunamadı"})
		return
	}

	if time.Now().After(otp.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kodun süresi dolmuş"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(otp.Code), []byte(req.Code)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kod hatalı"})
		return
	}

	otp.IsUsed = true
	config.DB.Save(&otp)

	var user models.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		user = models.User{Phone: req.Phone}
		config.DB.Create(&user)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Giriş başarılı", "user_id": user.ID})
}
