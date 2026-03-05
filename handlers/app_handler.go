package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"stajyer-backend/config"
	"stajyer-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateApplication: Başvuru oluşturur ve User tablosundaki profil bilgilerini senkronize eder
func CreateApplication(c *gin.Context) {
	// Form verilerini alıyoruz (React tarafındaki FormData isimleriyle uyumlu)
	phone := c.PostForm("phone")
	adSoyad := c.PostForm("full_name")
	email := c.PostForm("email")
	okul := c.PostForm("university")
	bolum := c.PostForm("department")
	stajTuru := c.PostForm("intern_type")

	// 1. Mükerrer Başvuru Kontrolü
	var existingApp models.Application
	if err := config.DB.Where("phone = ?", phone).First(&existingApp).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bu telefon numarası ile daha önce başvuru yapılmıştır."})
		return
	}

	// 2. Kullanıcı Kontrolü ve Profil Bilgilerini Güncelleme (Dataset Senkronizasyonu)
	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		// Kullanıcı yoksa oluştur (Auto-Register)
		user = models.User{
			Phone:              phone,
			FullNames:          adSoyad,
			Email:              email,
			University:         okul,
			Department:         bolum,
			IsProfileCompleted: true,
		}
		config.DB.Create(&user)
	} else {
		// Kullanıcı varsa, başvuru formundaki güncel bilgilerle User tablosunu güncelle
		config.DB.Model(&user).Updates(models.User{
			FullNames:          adSoyad,
			Email:              email,
			University:         okul,
			Department:         bolum,
			IsProfileCompleted: true,
		})
	}

	// 3. Dosya İşlemleri (CV ve Motivasyon Mektubu)
	cvFile, _ := c.FormFile("cv_file") // Frontend'deki name="cv_file"
	letterFile, _ := c.FormFile("letter_file")

	var cvUUID, cvPath, letterUUID, letterPath string

	if cvFile != nil {
		cvUUID = uuid.New().String()
		cvPath = filepath.ToSlash(filepath.Join("uploads", fmt.Sprintf("cv_%s%s", cvUUID, filepath.Ext(cvFile.Filename))))
		if err := c.SaveUploadedFile(cvFile, cvPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "CV kaydedilemedi"})
			return
		}
	}

	if letterFile != nil {
		letterUUID = uuid.New().String()
		letterPath = filepath.ToSlash(filepath.Join("uploads", fmt.Sprintf("letter_%s%s", letterUUID, filepath.Ext(letterFile.Filename))))
		c.SaveUploadedFile(letterFile, letterPath)
	}

	// 4. Başvuru Kaydı Oluşturma
	app := models.Application{
		FullNames:      adSoyad,
		Email:          email,
		Phone:          phone,
		University:     okul,
		Department:     bolum,
		InternshipType: stajTuru,
		CVUUID:         cvUUID,
		CVPath:         cvPath,
		LetterUUID:     letterUUID,
		LetterPath:     letterPath,
		UserID:         user.ID,
		Durum:          "Beklemede",
	}

	if err := config.DB.Create(&app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Başvuru veritabanına kaydedilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Başvuru başarıyla tamamlandı!",
		"status":  "Beklemede",
		"user_id": user.ID,
	})
}

// GetApplications: Admin paneli için tüm başvuruları getirir
func GetApplications(c *gin.Context) {
	var apps []models.Application
	if err := config.DB.Order("created_at desc").Find(&apps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Veriler çekilemedi"})
		return
	}
	c.JSON(http.StatusOK, apps)
}

// GetApplicationDetail: Tek bir başvurunun detayını getirir
func GetApplicationDetail(c *gin.Context) {
	id := c.Param("id")
	var application models.Application
	if err := config.DB.First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Başvuru bulunamadı"})
		return
	}
	c.JSON(http.StatusOK, application)
}

// UpdateApplicationStatus: Adminin onay/ret işlemini gerçekleştirir
func UpdateApplicationStatus(c *gin.Context) {
	id := c.Param("id")
	status := c.Query("status") // ?status=Onaylandı gibi

	if err := config.DB.Model(&models.Application{}).Where("id = ?", id).Update("durum", status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Durum güncellenemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Durum güncellendi", "new_status": status})
}

// DeleteApplication: Başvuruyu siler
func DeleteApplication(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Delete(&models.Application{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Silme hatası"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Başvuru silindi"})
}

// YENİ EKLEDİM
// CheckApplicationStatus: Kullanıcının başvuru durumunu telefon numarasına göre döndürür
func CheckApplicationStatus(c *gin.Context) {
	phone := c.Param("phone")
	var app models.Application

	if err := config.DB.Where("phone = ?", phone).First(&app).Error; err != nil {
		// Eğer başvuru bulunamazsa
		c.JSON(http.StatusOK, gin.H{
			"has_application": false,
			"status":          nil,
		})
		return
	}

	// Başvuru bulunursa
	c.JSON(http.StatusOK, gin.H{
		"has_application": true,
		"status":          app.Durum, // Veritabanındaki 'Beklemede' veya 'Onaylandı' durumu
	})
}
