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

// RequestOTP: Kullanıcı telefon girince OTP üretir, DB'ye kaydeder ve TERMİNALE YAZAR
func RequestOTP(c *gin.Context) {
	type Request struct {
		Phone string `json:"phone" binding:"required"`
	}

	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz istek"})
		return
	}

	// 1. Telefon format kontrolü (Utils üzerinden 5xx kuralı)
	if !utils.IsValidTurkishPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefon 10 hane olmalı ve 5 ile başlamalıdır"})
		return
	}

	// 2. Rate Limit Kontrolü (Aynı numaraya 60 saniyede bir kod)
	var lastOTP models.OTP
	err := config.DB.Where("phone = ?", req.Phone).Order("created_at DESC").First(&lastOTP).Error
	if err == nil && time.Since(lastOTP.CreatedAt) < 60*time.Second {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Lütfen yeni kod için 60 saniye bekleyin."})
		return
	}

	// 3. Kod Üretimi
	otpCode := utils.GenerateOTP()

	// --- TERMİNAL DESTEĞİ: Kodu burada konsola yazdırıyoruz ---
	fmt.Printf("\n==================================\n")
	fmt.Printf("🚀 YENİ OTP TALEBİ\n")
	fmt.Printf("📱 TELEFON: %s\n", req.Phone)
	fmt.Printf("🔑 ÜRETİLEN KOD: %s\n", otpCode)
	fmt.Printf("==================================\n\n")

	// 4. Güvenlik için kodu Hash'le ve DB'ye kaydet
	hashedCode, _ := bcrypt.GenerateFromPassword([]byte(otpCode), bcrypt.DefaultCost)

	otp := models.OTP{
		Phone:     req.Phone,
		Code:      string(hashedCode),
		ExpiresAt: time.Now().Add(2 * time.Minute), // 2 dakika geçerli
		IsUsed:    false,
	}

	if err := config.DB.Create(&otp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OTP veritabanına kaydedilemedi"})
		return
	}

	// Yanıt gönder (Güvenlik için 'otp': otpCode kısmı canlıda kaldırılmalı)
	c.JSON(http.StatusOK, gin.H{"message": "Doğrulama kodu oluşturuldu.", "otp": otpCode})
}

// VerifyOTP: Gelen kodu doğrular ve kullanıcıyı giriş yaptırır
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

	// 1. En güncel ve kullanılmamış kodu bul
	var otp models.OTP
	err := config.DB.Where("phone = ? AND is_used = ?", req.Phone, false).Order("created_at DESC").First(&otp).Error
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Geçerli bir doğrulama kodu bulunamadı"})
		return
	}

	// 2. Süre kontrolü
	if time.Now().After(otp.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Kodun süresi dolmuş, lütfen yeni kod isteyin"})
		return
	}

	// 3. Hash kıyaslaması (Bcrypt)
	if err := bcrypt.CompareHashAndPassword([]byte(otp.Code), []byte(req.Code)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Girdiğiniz kod hatalı"})
		return
	}

	// 4. Kod kullanıldı olarak işaretle
	otp.IsUsed = true
	config.DB.Save(&otp)

	// 5. Kullanıcıyı bul veya yoksa oluştur (Auto-Register)
	var user models.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		user = models.User{Phone: req.Phone}
		config.DB.Create(&user)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Giriş başarılı!",
		"user_id": user.ID,
		"phone":   user.Phone,
	})
}
