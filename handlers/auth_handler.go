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
	"gorm.io/gorm"
)

// RequestOTP: Herkes için (Şirket, Öğrenci, Admin) giriş kapısı. Başvuru zorunluluğu KALDIRILDI.
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

	// SPAM KONTROLÜ
	var lastOTP models.OTP
	err := config.DB.Where("phone = ?", req.Phone).Order("created_at DESC").First(&lastOTP).Error
	if err == nil && time.Since(lastOTP.CreatedAt) < 60*time.Second {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Lütfen yeni kod için 60 saniye bekleyin."})
		return
	}

	// OTP ÜRETME
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

// VerifyOTP: Kod doğrulama sonrası kullanıcıyı sisteme alır ve rolünü bildirir
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

	// Kullanıcı yoksa varsayılan olarak "student" oluştur
	if err := config.DB.Where("phone = ?", req.Phone).Preload("StudentProfile").Preload("CompanyProfile").First(&user).Error; err != nil {
		user = models.User{Phone: req.Phone, Role: "student"} // Şirketleri admin panelinden ekleyebiliriz
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

// CompleteProfile: YENİ MİMARİ. Kullanıcının rolüne göre alt tabloları doldurur.
func CompleteProfile(c *gin.Context) {
	type ProfileUpdateRequest struct {
		Phone string `json:"phone" binding:"required"`
		// Ortak
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Role     string `json:"role"` // Frontend'den form tipini (öğrenci mi şirket mi) anlamak için

		// Sadece Öğrenci
		TCKimlik   string `json:"tc_kimlik"`
		University string `json:"university"`
		Department string `json:"department"`

		// Sadece Şirket
		CompanyName string `json:"company_name"`
		TaxNumber   string `json:"tax_number"`
		Industry    string `json:"industry"`
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

	// Ana kullanıcı tablosunu güncelle
	user.FullNames = req.FullName
	user.Email = req.Email
	user.IsProfileCompleted = true

	// Kullanıcının rolünü güncelle (Eğer öğrenci/şirket seçim ekranından geliyorsa)
	if req.Role != "" {
		user.Role = req.Role
	}
	config.DB.Save(&user)

	// Alt tablolara verileri yaz
	if user.Role == "student" {
		var student models.Student
		if err := config.DB.Where("user_id = ?", user.ID).First(&student).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				student = models.Student{
					UserID:     user.ID,
					TCNo:       req.TCKimlik,
					University: req.University,
					Department: req.Department,
				}
				config.DB.Create(&student)
			}
		} else {
			student.TCNo = req.TCKimlik
			student.University = req.University
			student.Department = req.Department
			config.DB.Save(&student)
		}
	} else if user.Role == "company" {
		var company models.Company
		if err := config.DB.Where("user_id = ?", user.ID).First(&company).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				company = models.Company{
					UserID:      user.ID,
					CompanyName: req.CompanyName,
					TaxNumber:   req.TaxNumber,
					Industry:    req.Industry,
				}
				config.DB.Create(&company)
			}
		} else {
			company.CompanyName = req.CompanyName
			company.TaxNumber = req.TaxNumber
			company.Industry = req.Industry
			config.DB.Save(&company)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profil güncellendi"})
}

// GetUserByPhone: Kullanıcının genel ve rolüne ait detaylarını getirir
func GetUserByPhone(c *gin.Context) {
	phone := c.Param("phone")
	var user models.User

	// Preload ile ilişkili tabloları (öğrenci/şirket detaylarını) da çekiyoruz
	if err := config.DB.Where("phone = ?", phone).Preload("StudentProfile").Preload("CompanyProfile").First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	c.JSON(http.StatusOK, user)
}
