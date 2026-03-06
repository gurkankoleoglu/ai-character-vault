package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"stajyer-backend/config"
	"stajyer-backend/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateApplication: Yeni başvuru oluşturur ve User tablosunu günceller
func CreateApplication(c *gin.Context) {
	phone := c.PostForm("phone")
	adSoyad := c.PostForm("full_name")
	email := c.PostForm("email")
	okul := c.PostForm("university")
	bolum := c.PostForm("department")
	stajTuru := c.PostForm("intern_type")
	stajTarihleri := c.PostForm("stajTarihleri")

	// 1. Mükerrer Başvuru Kontrolü (Beklemede veya Onaylı varsa engelle)
	var existingApp models.Application
	if err := config.DB.Where("phone = ? AND durum IN ?", phone, []string{"Beklemede", "Onaylandı"}).First(&existingApp).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Devam eden veya onaylanmış bir başvurunuz zaten var."})
		return
	}

	// 2. Kullanıcı Bilgilerini Güncelle (Dashboard için Senkronizasyon)
	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err == nil {
		config.DB.Model(&user).Updates(models.User{
			FullNames:          adSoyad,
			Email:              email,
			University:         okul,
			Department:         bolum,
			IsProfileCompleted: true,
		})
	}

	// 3. Dosya Kaydetme (CV ve Mektup)
	cvFile, _ := c.FormFile("cv_file")
	letterFile, _ := c.FormFile("letter_file")
	var cvUUID, cvPath, letterUUID, letterPath string

	if cvFile != nil {
		cvUUID = uuid.New().String()
		cvPath = filepath.ToSlash(filepath.Join("uploads", fmt.Sprintf("cv_%s%s", cvUUID, filepath.Ext(cvFile.Filename))))
		c.SaveUploadedFile(cvFile, cvPath)
	}
	if letterFile != nil {
		letterUUID = uuid.New().String()
		letterPath = filepath.ToSlash(filepath.Join("uploads", fmt.Sprintf("letter_%s%s", letterUUID, filepath.Ext(letterFile.Filename))))
		c.SaveUploadedFile(letterFile, letterPath)
	}

	// 4. Başvuruyu Kaydet
	app := models.Application{
		FullNames:      adSoyad,
		Email:          email,
		Phone:          phone,
		University:     okul,
		Department:     bolum,
		InternshipType: stajTuru,
		StajTarihleri:  stajTarihleri,
		CVUUID:         cvUUID,
		CVPath:         cvPath,
		LetterUUID:     letterUUID,
		LetterPath:     letterPath,
		UserID:         user.ID,
		Durum:          "Beklemede",
	}

	if err := config.DB.Create(&app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Veritabanı hatası"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Başvurunuz alındı, onay bekleniyor.", "status": "Beklemede"})
}

// UpdateApplicationStatus: Admin Onay/Red yaptığında çalışır ve KONSOLA SMS BASAR
func UpdateApplicationStatus(c *gin.Context) {
	id := c.Param("id")
	status := c.Query("status") // Onaylandı veya Reddedildi

	var application models.Application
	if err := config.DB.First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Başvuru bulunamadı"})
		return
	}

	// Veritabanı Güncelleme
	if err := config.DB.Model(&application).Update("durum", status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Durum güncellenemedi"})
		return
	}

	// 📱 SMS SİMÜLATÖRÜ - TERMİNAL ÇIKTISI (Switch yapısı ile optimize edildi)
	fmt.Println("\n" + strings.Repeat("*", 60))
	fmt.Printf("🚀 [SMS SİMÜLATÖRÜ] - SAAT: %s\n", time.Now().Format("15:04:05"))
	fmt.Printf("GÖNDERİLEN NUMARA: %s\n", application.Phone)

	switch status {
	case "Onaylandı":
		fmt.Printf("MESAJ: Sayın %s, staj başvurunuz ONAYLANMIŞTIR. Artık giriş yapabilirsiniz. ✅\n", application.FullNames)
	case "Reddedildi":
		fmt.Printf("MESAJ: Sayın %s, staj başvurunuz maalesef REDDEDİLMİŞTİR. Giriş yetkiniz bulunmamaktadır. ❌\n", application.FullNames)
	default:
		fmt.Printf("MESAJ: Sayın %s, başvurunuzun durumu güncellendi: %s\n", application.FullNames, status)
	}
	fmt.Println(strings.Repeat("*", 60) + "\n")

	c.JSON(http.StatusOK, gin.H{"message": "Durum güncellendi ve kullanıcı bilgilendirildi.", "new_status": status})
}

// GetApplications: Tüm başvuruları listeler
func GetApplications(c *gin.Context) {
	var apps []models.Application
	config.DB.Order("created_at desc").Find(&apps)
	c.JSON(http.StatusOK, apps)
}

// GetApplicationDetail: Detay getirir
func GetApplicationDetail(c *gin.Context) {
	id := c.Param("id")
	var app models.Application
	if err := config.DB.First(&app, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bulunamadı"})
		return
	}
	c.JSON(http.StatusOK, app)
}

// DeleteApplication: Başvuruyu siler
func DeleteApplication(c *gin.Context) {
	id := c.Param("id")
	config.DB.Delete(&models.Application{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Silindi"})
}

// CheckApplicationStatus: Telefon ile durum sorgular
func CheckApplicationStatus(c *gin.Context) {
	phone := c.Param("phone")
	var app models.Application
	if err := config.DB.Where("phone = ?", phone).Order("created_at desc").First(&app).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"has_application": false})
		return
	}
	c.JSON(http.StatusOK, gin.H{"has_application": true, "status": app.Durum})
}
