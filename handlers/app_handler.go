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

// CreateApplication: Başvuru oluşturur ve mükerrer kaydı engeller
func CreateApplication(c *gin.Context) {
	phone := c.PostForm("telefon")
	adSoyad := c.PostForm("adSoyad")
	email := c.PostForm("email")

	// 1. Mükerrer Başvuru Kontrolü
	var existingApp models.Application
	if err := config.DB.Where("phone = ?", phone).First(&existingApp).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bu telefon numarası ile daha önce başvuru yapılmıştır."})
		return
	}

	// 2. Kullanıcı Kontrolü ve Otomatik Kayıt
	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		user = models.User{
			Phone:     phone,
			FullNames: adSoyad,
			Email:     email,
		}
		config.DB.Create(&user)
	}

	// 3. Dosya İşlemleri (UUID)
	cvFile, _ := c.FormFile("dosya")
	letterFile, _ := c.FormFile("motivasyonDosya")
	var cvUUID, cvPath, letterUUID, letterPath string

	if cvFile != nil {
		cvUUID = uuid.New().String()
		// URL uyumluluğu için ToSlash kullanıyoruz
		cvPath = filepath.ToSlash(filepath.Join("uploads", fmt.Sprintf("cv_%s%s", cvUUID, filepath.Ext(cvFile.Filename))))
		c.SaveUploadedFile(cvFile, cvPath)
	}

	if letterFile != nil {
		letterUUID = uuid.New().String()
		letterPath = filepath.ToSlash(filepath.Join("uploads", fmt.Sprintf("letter_%s%s", letterUUID, filepath.Ext(letterFile.Filename))))
		c.SaveUploadedFile(letterFile, letterPath)
	}

	// 4. Başvuru Kaydı
	app := models.Application{
		FullNames:      adSoyad,
		Email:          email,
		Phone:          phone,
		University:     c.PostForm("okul"),
		Department:     c.PostForm("bolum"),
		InternshipType: c.PostForm("stajTuru"),
		CVUUID:         cvUUID,
		CVPath:         cvPath, // Frontend'de "i" butonu ile erişilecek yol
		LetterUUID:     letterUUID,
		LetterPath:     letterPath, // Frontend'de "i" butonu ile erişilecek yol
		UserID:         user.ID,
		Durum:          "Beklemede", // Varsayılan durum
	}

	if err := config.DB.Create(&app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Veritabanı hatası"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Başvuru başarıyla tamamlandı!", "user_id": user.ID})
}

// GetApplications: Tüm başvuruları listeler (Admin Panel için)
func GetApplications(c *gin.Context) {
	var apps []models.Application
	if err := config.DB.Order("created_at desc").Find(&apps).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Veriler çekilemedi"})
		return
	}
	c.JSON(http.StatusOK, apps)
}

// UpdateApplicationStatus: Başvurunun durumunu (Onaylandı/Reddedildi) günceller
func UpdateApplicationStatus(c *gin.Context) {
	id := c.Param("id")
	status := c.Query("status") //

	var app models.Application
	if err := config.DB.First(&app, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Başvuru bulunamadı"})
		return
	}

	app.Durum = status
	if err := config.DB.Save(&app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Durum güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Durum güncellendi", "status": app.Durum})
}

// UpdateApplication: Başvuru bilgilerini düzenler (API Düzeyinde Aktif)
func UpdateApplication(c *gin.Context) {
	id := c.Param("id")
	var app models.Application

	if err := config.DB.First(&app, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Başvuru bulunamadı"})
		return
	}

	app.FullNames = c.PostForm("adSoyad")
	app.Email = c.PostForm("email")
	app.University = c.PostForm("okul")
	app.Department = c.PostForm("bolum")
	app.InternshipType = c.PostForm("stajTuru")
	app.Phone = c.PostForm("telefon")

	if err := config.DB.Save(&app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Güncelleme başarısız"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Başvuru güncellendi", "data": app})
}

// DeleteApplication: Başvuruyu veritabanından siler (API Düzeyinde Aktif)
func DeleteApplication(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Delete(&models.Application{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Silme işlemi başarısız"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Başvuru başarıyla silindi"})
}
