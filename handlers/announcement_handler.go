package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"stajyer-backend/config"
	"stajyer-backend/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// InitializeDatabase: Veritabanı tablolarını oluşturur
func InitializeDatabase() {
	err := config.DB.AutoMigrate(&models.Announcement{})
	if err != nil {
		panic("Announcement tablosu oluşturulurken hata oluştu: " + err.Error())
	}
}

// CreateAnnouncement: Yeni bir duyuru oluşturur
func CreateAnnouncement(c *gin.Context) {
	// JSON yerine FormData'dan verileri çekiyoruz
	title := c.PostForm("title")
	annType := c.PostForm("type")
	subject := c.PostForm("subject")
	content := c.PostForm("content")
	dateStr := c.PostForm("date")

	parsedDate := time.Now()
	if pd, err := time.Parse(time.RFC3339, dateStr); err == nil {
		parsedDate = pd
	}

	// Resim dosyasını al ve uploads klasörüne kaydet
	var imageURL string
	file, _ := c.FormFile("image")
	if file != nil {
		fileUUID := uuid.New().String()
		ext := filepath.Ext(file.Filename)
		fileName := fmt.Sprintf("ann_%s%s", fileUUID, ext)
		savePath := filepath.ToSlash(filepath.Join("uploads", fileName))

		if err := c.SaveUploadedFile(file, savePath); err == nil {
			imageURL = "http://localhost:8080/" + savePath
		}
	}

	announcement := models.Announcement{
		Title:    title,
		Type:     annType,
		Subject:  subject,
		Content:  content,
		Date:     parsedDate,
		ImageURL: imageURL,
	}

	// Save the announcement to the database
	if err := config.DB.Create(&announcement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Duyuru oluşturulamadı", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Duyuru başarıyla oluşturuldu", "data": announcement})
}

// GetAnnouncements: Tüm duyuruları getirir
func GetAnnouncements(c *gin.Context) {
	var announcements []models.Announcement
	if err := config.DB.Order("created_at desc").Find(&announcements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Duyurular getirilemedi"})
		return
	}
	c.JSON(http.StatusOK, announcements)
}

// GetAnnouncementDetail: Belirli bir duyuruyu getirir
func GetAnnouncementDetail(c *gin.Context) {
	id := c.Param("id")
	var announcement models.Announcement

	if err := config.DB.First(&announcement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Duyuru bulunamadı"})
		return
	}
	c.JSON(http.StatusOK, announcement)
}

// UpdateAnnouncement: Belirli bir duyuruyu günceller
func UpdateAnnouncement(c *gin.Context) {
	id := c.Param("id")
	var announcement models.Announcement

	if err := config.DB.First(&announcement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Duyuru bulunamadı"})
		return
	}

	// Yeni verileri ata
	announcement.Title = c.PostForm("title")
	announcement.Type = c.PostForm("type")
	announcement.Subject = c.PostForm("subject")
	announcement.Content = c.PostForm("content")

	dateStr := c.PostForm("date")
	if pd, err := time.Parse(time.RFC3339, dateStr); err == nil {
		announcement.Date = pd
	}

	// Eğer güncelleme sırasında yeni bir resim yüklendiyse onu değiştir
	file, _ := c.FormFile("image")
	if file != nil {
		fileUUID := uuid.New().String()
		ext := filepath.Ext(file.Filename)
		fileName := fmt.Sprintf("ann_%s%s", fileUUID, ext)
		savePath := filepath.ToSlash(filepath.Join("uploads", fileName))

		if err := c.SaveUploadedFile(file, savePath); err == nil {
			announcement.ImageURL = "http://localhost:8080/" + savePath
		}
	}

	if err := config.DB.Save(&announcement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Duyuru güncellenemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Duyuru başarıyla güncellendi", "data": announcement})
}

// DeleteAnnouncement: Belirli bir duyuruyu siler
func DeleteAnnouncement(c *gin.Context) {
	id := c.Param("id")

	// Validate ID
	announcementID, err := strconv.Atoi(id)
	if err != nil || announcementID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz duyuru ID"})
		return
	}

	// Check if the announcement exists
	var announcement models.Announcement
	if err := config.DB.First(&announcement, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Duyuru bulunamadı"})
		return
	}

	// Delete the announcement
	if err := config.DB.Delete(&announcement).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Duyuru silinemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Duyuru başarıyla silindi"})
}
