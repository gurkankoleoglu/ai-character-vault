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

func CreateApplication(c *gin.Context) {
	// 1. Form ve Kullanıcı Bilgisini Al
	// Frontend'den gelen user_id'yi alıyoruz
	userID := c.PostForm("user_id")

	// 2. Dosyaları İşle
	cvFile, _ := c.FormFile("dosya")
	letterFile, _ := c.FormFile("motivasyonDosya")

	var cvUUID, cvPath, letterUUID, letterPath string

	if cvFile != nil {
		cvUUID = uuid.New().String() // Benzersiz UUID oluştur
		// Örn: uploads/cv_6f9a2...png
		cvPath = fmt.Sprintf("uploads/cv_%s%s", cvUUID, filepath.Ext(cvFile.Filename))
		c.SaveUploadedFile(cvFile, cvPath)
	}

	if letterFile != nil {
		letterUUID = uuid.New().String()
		letterPath = fmt.Sprintf("uploads/letter_%s%s", letterUUID, filepath.Ext(letterFile.Filename))
		c.SaveUploadedFile(letterFile, letterPath)
	}

	// 3. Veritabanı Kaydı
	var uID uint
	fmt.Sscanf(userID, "%d", &uID)

	app := models.Application{
		FullNames:      c.PostForm("adSoyad"),
		Email:          c.PostForm("email"),
		Phone:          c.PostForm("telefon"),
		University:     c.PostForm("okul"),
		Department:     c.PostForm("bolum"),
		InternshipType: c.PostForm("stajTuru"),
		CVUUID:         cvUUID,
		CVPath:         cvPath,
		LetterUUID:     letterUUID,
		LetterPath:     letterPath,
		UserID:         uID, // Giriş yapan kullanıcının ID'si
	}

	if err := config.DB.Create(&app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Veritabanı hatası"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Başvuru ve dosyalar başarıyla kaydedildi!"})
}
