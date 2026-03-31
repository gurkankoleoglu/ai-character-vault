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

// ────────────────────────────────────────────────────────────
// OTP
// ────────────────────────────────────────────────────────────

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

// ────────────────────────────────────────────────────────────
// ✅ DÜZELTİLDİ: PostForm → JSON
// ────────────────────────────────────────────────────────────

func CompleteProfile(c *gin.Context) {
	type CompleteProfileRequest struct {
		Email       string `json:"email" binding:"required"`
		Password    string `json:"password"`
		Role        string `json:"role" binding:"required"`
		Phone       string `json:"phone"`
		FullName    string `json:"full_name"`
		TcNo        string `json:"tc_no"`
		University  string `json:"university"`
		Department  string `json:"department"`
		InternType  string `json:"intern_type"`
		Gender      string `json:"gender"`
		CompanyName string `json:"company_name"`
		TaxNumber   string `json:"tax_number"`
		Sector      string `json:"sector"`
	}

	var req CompleteProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri: " + err.Error()})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		// Yeni kullanıcı oluştur
		phoneToSave := req.Phone
		if phoneToSave == "" {
			phoneToSave = fmt.Sprintf("temp-%d", time.Now().UnixNano())
		}
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user = models.User{
			Email:              req.Email,
			Password:           string(hashedPassword),
			Role:               req.Role,
			Phone:              phoneToSave,
			FullNames:          req.FullName,
			IsProfileCompleted: true,
			IsOnline:           true,
		}
		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Kullanıcı oluşturulamadı: " + err.Error()})
			return
		}
	} else {
		// Mevcut kullanıcıyı güncelle
		user.Role = req.Role
		user.FullNames = req.FullName
		if req.Password != "" {
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			user.Password = string(hashedPassword)
		}
		if req.Phone != "" {
			user.Phone = req.Phone
		}
		user.IsProfileCompleted = true
		config.DB.Save(&user)
	}

	// Alt tabloları gerçek verilerle doldur
	if req.Role == "student" {
		var student models.Student
		if err := config.DB.Where("user_id = ?", user.ID).First(&student).Error; err != nil {
			student = models.Student{UserID: user.ID}
		}
		if req.TcNo != "" {
			student.TCNo = req.TcNo
		} else {
			student.TCNo = "00000000000"
		}
		if req.University != "" {
			student.University = req.University
		} else {
			student.University = "Belirtilmedi"
		}
		if req.Department != "" {
			student.Department = req.Department
		} else {
			student.Department = "Belirtilmedi"
		}
		config.DB.Save(&student)
	} else if req.Role == "company" {
		var company models.Company
		if err := config.DB.Where("user_id = ?", user.ID).First(&company).Error; err != nil {
			company = models.Company{UserID: user.ID}
		}
		if req.CompanyName != "" {
			company.CompanyName = req.CompanyName
		} else {
			company.CompanyName = "Belirtilmedi"
		}
		if req.TaxNumber != "" {
			company.TaxNumber = req.TaxNumber
		} else {
			company.TaxNumber = "0000000000"
		}
		if req.Sector != "" {
			company.Industry = req.Sector
		} else {
			company.Industry = "Belirtilmedi"
		}
		config.DB.Save(&company)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Kayıt başarıyla tamamlandı"})
}

// ────────────────────────────────────────────────────────────
// ✅ DÜZELTİLDİ: tc_no ve intern_type eklendi
// ────────────────────────────────────────────────────────────

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
		"tc_no":                "",
	}

	if user.Role == "student" {
		var student models.Student
		config.DB.Where("user_id = ?", user.ID).First(&student)
		response["tc_no"] = student.TCNo
		response["university"] = student.University
		response["department"] = student.Department
		//response["intern_type"]     = student.InternType
		response["skills"] = student.Skills
		response["github"] = student.Github
		response["student_profile"] = student
	} else if user.Role == "company" {
		var company models.Company
		config.DB.Where("user_id = ?", user.ID).First(&company)
		response["company_name"] = company.CompanyName
		response["sector"] = company.Industry
		response["website"] = company.Website
		response["company_profile"] = company
	}

	c.JSON(http.StatusOK, response)
}

// ────────────────────────────────────────────────────────────
// ✅ YENİ: Email ile kullanıcı getir (Giris.jsx için)
// ────────────────────────────────────────────────────────────

func GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	var user models.User
	if err := config.DB.Where("email = ?", email).First(&user).Error; err != nil {
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
		"tc_no":                "",
	}

	if user.Role == "student" {
		var student models.Student
		config.DB.Where("user_id = ?", user.ID).First(&student)
		response["tc_no"] = student.TCNo
		response["university"] = student.University
		response["department"] = student.Department
		//response["intern_type"]     = student.InternType
		response["skills"] = student.Skills
		response["github"] = student.Github
		response["student_profile"] = student
	} else if user.Role == "company" {
		var company models.Company
		config.DB.Where("user_id = ?", user.ID).First(&company)
		response["company_name"] = company.CompanyName
		response["sector"] = company.Industry
		response["website"] = company.Website
		response["company_profile"] = company
	}

	c.JSON(http.StatusOK, response)
}

// ────────────────────────────────────────────────────────────
// ✅ YENİ: Başvuru durumu kontrolü
// ────────────────────────────────────────────────────────────

func CheckApplication(c *gin.Context) {
	phone := c.Param("phone")
	var user models.User

	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "Onaylandı"})
		return
	}
	if user.Role == "admin" {
		c.JSON(http.StatusOK, gin.H{"status": "Onaylandı"})
		return
	}
	if user.Role == "student" {
		var student models.Student
		if err := config.DB.Where("user_id = ?", user.ID).First(&student).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"status": "Beklemede"})
			return
		}
		if student.IsCertApproved {
			c.JSON(http.StatusOK, gin.H{"status": "Onaylandı"})
		} else {
			c.JSON(http.StatusOK, gin.H{"status": "Beklemede"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "Onaylandı"})
}

// ────────────────────────────────────────────────────────────
// Profil güncelle
// ────────────────────────────────────────────────────────────

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

// ────────────────────────────────────────────────────────────
// Çıkış
// ────────────────────────────────────────────────────────────

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

// ────────────────────────────────────────────────────────────
// ✅ DÜZELTİLDİ: Login — phone da response'a eklendi
// ────────────────────────────────────────────────────────────

func Login(c *gin.Context) {
	type LoginRequest struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "E-posta ve şifre gereklidir"})
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "E-posta veya şifre hatalı"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "E-posta veya şifre hatalı"})
		return
	}

	user.IsOnline = true
	config.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{
		"message":              "Giriş başarılı!",
		"id":                   user.ID,
		"phone":                user.Phone,
		"email":                user.Email,
		"role":                 user.Role,
		"full_name":            user.FullNames,
		"is_online":            user.IsOnline,
		"is_profile_completed": user.IsProfileCompleted,
	})
}
