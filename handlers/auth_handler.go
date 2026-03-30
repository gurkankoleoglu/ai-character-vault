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
	"golang.org/x/crypto/bcrypt"
)

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

	var lastOTP models.OTP
	err := config.DB.Where("phone = ?", req.Phone).Order("created_at DESC").First(&lastOTP).Error
	if err == nil && time.Since(lastOTP.CreatedAt) < 60*time.Second {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Lütfen yeni kod için 60 saniye bekleyin."})
		return
	}

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

	if err := config.DB.Where("phone = ?", req.Phone).Preload("StudentProfile").Preload("CompanyProfile").First(&user).Error; err != nil {
		user = models.User{Phone: req.Phone, Role: "student", IsOnline: true}
		config.DB.Create(&user)
		isNewUser = true
	} else {
		user.IsOnline = true
		config.DB.Save(&user)
		if user.FullNames == "" {
			isNewUser = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Giriş başarılı!",
		"user_id":     user.ID,
		"phone":       user.Phone,
		"role":        user.Role,
		"full_names":  user.FullNames,
		"is_online":   user.IsOnline,
		"is_new_user": isNewUser,
	})
}

func CompleteProfile(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	role := c.PostForm("role")
	phoneForm := c.PostForm("phone") // Formdan gelen gerçek veri

	if email == "" || role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "E-posta ve rol bilgisi zorunludur"})
		return
	}

	var user models.User
	// E-posta ile kullanıcıyı bul veya yeni oluştur
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {

		// KRİTİK DÜZELTME: Telefon unique (benzersiz) olduğu için, boş geldiğinde hata fırlatmasını
		// engellemek amacıyla zaman damgalı benzersiz bir "sahte" numara atıyoruz.
		phoneToSave := phoneForm
		if phoneToSave == "" {
			phoneToSave = fmt.Sprintf("temp-%d", time.Now().UnixNano())
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		user = models.User{
			Email:              email,
			Password:           string(hashedPassword),
			Role:               role,
			Phone:              phoneToSave, // Artık her kayıt için tamamen benzersiz
			IsProfileCompleted: true,
			IsOnline:           true,
		}

		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı hesabı oluşturulamadı: " + err.Error()})
			return
		}
	} else {
		// Mevcut Kullanıcı Güncelleme İşlemi
		user.Role = role
		if password != "" {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			user.Password = string(hashedPassword)
		}
		// Sadece formdan GERÇEK bir numara edildiyse üzerine yaz, sahte olanla değiştirme
		if phoneForm != "" {
			user.Phone = phoneForm
		}
		user.IsProfileCompleted = true
		config.DB.Save(&user)
	}

	// Profile göre alt tablo kayıtlarını oluşturma (NOT NULL kısıtlamalarını aşmak için varsayılan değerler)
	if role == "student" {
		var student models.Student
		if err := config.DB.Where("user_id = ?", user.ID).First(&student).Error; err != nil {
			student = models.Student{
				UserID:     user.ID,
				TCNo:       "00000000000",
				University: "Belirtilmedi",
				Department: "Belirtilmedi",
			}
			config.DB.Create(&student)
		}
	} else if role == "company" {
		var company models.Company
		if err := config.DB.Where("user_id = ?", user.ID).First(&company).Error; err != nil {
			company = models.Company{
				UserID:      user.ID,
				CompanyName: "Belirtilmedi",
				TaxNumber:   "0000000000",
				Industry:    "Belirtilmedi",
			}
			config.DB.Create(&company)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kayıt başarıyla tamamlandı"})
}

func GetUserByPhone(c *gin.Context) {
	phone := c.Param("phone")
	var user models.User

	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	response := gin.H{
		"id":                   user.ID,
		"phone":                user.Phone,
		"email":                user.Email,
		"full_name":            user.FullNames,
		"role":                 user.Role,
		"location":             user.Location,
		"bio":                  user.Bio,
		"avatar_url":           user.AvatarURL,
		"is_profile_completed": user.IsProfileCompleted,
		"is_online":            user.IsOnline,
	}

	if user.Role == "student" {
		var student models.Student
		config.DB.Where("user_id = ?", user.ID).First(&student)
		response["student_profile"] = student
		response["university"] = student.University
		response["department"] = student.Department
		response["skills"] = student.Skills
		response["github"] = student.Github
	} else if user.Role == "company" {
		var company models.Company
		config.DB.Where("user_id = ?", user.ID).First(&company)
		response["company_profile"] = company
		response["company_name"] = company.CompanyName
		response["sector"] = company.Industry
		response["website"] = company.Website
	}

	c.JSON(http.StatusOK, response)
}

func UpdateUserProfile(c *gin.Context) {
	phone := c.Param("phone")

	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	user.FullNames = c.PostForm("full_name")
	user.Email = c.PostForm("email")
	user.Location = c.PostForm("location")
	user.Bio = c.PostForm("bio")

	avatarFile, _ := c.FormFile("avatar")
	if avatarFile != nil {
		ext := filepath.Ext(avatarFile.Filename)
		avatarPath := filepath.ToSlash(filepath.Join("uploads", fmt.Sprintf("avatar_%s%s", phone, ext)))
		if err := c.SaveUploadedFile(avatarFile, avatarPath); err == nil {
			user.AvatarURL = "http://localhost:8080/" + avatarPath
		}
	}
	config.DB.Save(&user)

	if user.Role == "student" {
		var student models.Student
		config.DB.Where("user_id = ?", user.ID).FirstOrCreate(&student, models.Student{UserID: user.ID})
		student.University = c.PostForm("university")
		student.Department = c.PostForm("department")
		student.Skills = c.PostForm("skills")
		student.Github = c.PostForm("github")
		config.DB.Save(&student)
	} else if user.Role == "company" {
		var company models.Company
		config.DB.Where("user_id = ?", user.ID).FirstOrCreate(&company, models.Company{UserID: user.ID})
		company.CompanyName = c.PostForm("company_name")
		company.Industry = c.PostForm("sector")
		company.Website = c.PostForm("website")
		config.DB.Save(&company)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Profil başarıyla güncellendi",
		"avatar_url": user.AvatarURL,
	})
}

func Logout(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefon numarası gerekli"})
		return
	}

	if err := config.DB.Model(&models.User{}).Where("phone = ?", req.Phone).Update("is_online", false).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Çıkış işlemi başarısız"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Başarıyla çıkış yapıldı"})
}
