package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"stajyer-backend/config"
	"stajyer-backend/models"
	"stajyer-backend/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// RequestOTP: Herkes için (Şirket, Öğrenci, Admin) giriş kapısı.
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
		user = models.User{Phone: req.Phone, Role: "student"}
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

// CompleteProfile: YENİ MİMARİ. Kullanıcının rolüne göre alt tabloları doldurur (Dosya Yükleme ve Upsert Destekli)
func CompleteProfile(c *gin.Context) {
	phone := c.PostForm("phone")
	role := c.PostForm("role")
	email := c.PostForm("email")

	if phone == "" || role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefon ve rol bilgisi zorunludur"})
		return
	}

	// 1. KULLANICIYI BUL VEYA SIFIRDAN YARAT (Upsert Mantığı)
	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		// Veritabanında yoksa sıfırdan oluştur
		user = models.User{
			Phone:              phone,
			Role:               role,
			Email:              email,
			IsProfileCompleted: true,
		}
		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı hesabı oluşturulamadı"})
			return
		}
	} else {
		// Varsa bilgilerini güncelle
		user.Email = email
		user.Role = role
		user.IsProfileCompleted = true
		config.DB.Save(&user)
	}

	// 2. ROL BAZLI ALT TABLOLARI DOLDUR
	// Rol Öğrenci ise:
	if role == "student" {
		user.FullNames = c.PostForm("full_name")
		config.DB.Save(&user) // Öğrencinin adını ana tabloya da kaydet

		// Dosya Yükleme İşlemi (Öğrenci Belgesi)
		file, _ := c.FormFile("student_document")
		var filePath string
		if file != nil {
			fileUUID := uuid.New().String()
			filePath = filepath.ToSlash(filepath.Join("uploads", fmt.Sprintf("cert_%s%s", fileUUID, filepath.Ext(file.Filename))))
			c.SaveUploadedFile(file, filePath)
		}

		var student models.Student
		if err := config.DB.Where("user_id = ?", user.ID).First(&student).Error; err != nil {
			// Öğrenci profili yoksa yarat
			student = models.Student{
				UserID:         user.ID,
				TCNo:           c.PostForm("tc_no"),
				University:     c.PostForm("university"),
				Department:     c.PostForm("department"),
				StudentCertUrl: filePath, // Yüklenen dosyanın yolu
			}
			config.DB.Create(&student)
		} else {
			// Öğrenci profili varsa güncelle
			student.TCNo = c.PostForm("tc_no")
			student.University = c.PostForm("university")
			student.Department = c.PostForm("department")
			if filePath != "" {
				student.StudentCertUrl = filePath // Sadece yeni dosya yüklendiyse güncelle
			}
			config.DB.Save(&student)
		}

		// Rol Şirket ise:
	} else if role == "company" {
		var company models.Company
		if err := config.DB.Where("user_id = ?", user.ID).First(&company).Error; err != nil {
			company = models.Company{
				UserID:      user.ID,
				CompanyName: c.PostForm("company_name"),
				TaxNumber:   c.PostForm("tax_number"),
				Industry:    c.PostForm("sector"), // Frontend'den gelen 'sector' verisi
			}
			config.DB.Create(&company)
		} else {
			company.CompanyName = c.PostForm("company_name")
			company.TaxNumber = c.PostForm("tax_number")
			company.Industry = c.PostForm("sector")
			config.DB.Save(&company)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profil başarıyla güncellendi"})
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
