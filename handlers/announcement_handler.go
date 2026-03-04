package handlers

import (
	"net/http"
	"stajyer-backend/config"
	"stajyer-backend/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InitializeDatabase: Veritabanı tablolarını oluşturur
func InitializeDatabase() {
	// Announcement tablosunu oluştur
	err := config.DB.AutoMigrate(&models.Announcement{})
	if err != nil {
		panic("Announcement tablosu oluşturulurken hata oluştu: " + err.Error())
	}
}

// CreateAnnouncement: Yeni bir duyuru oluşturur
func CreateAnnouncement(c *gin.Context) {
	var announcement models.Announcement

	// Bind the JSON data to the Announcement struct directly (only once)
	if err := c.ShouldBindJSON(&announcement); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz duyuru verisi", "details": err.Error()})
		return
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
	// Son eklenenlerin en üstte gelmesi için Order eklendi
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

	var updatedData models.Announcement
	if err := c.ShouldBindJSON(&updatedData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz duyuru verisi"})
		return
	}

	// Update fields
	announcement.Title = updatedData.Title
	announcement.Content = updatedData.Content
	announcement.Subject = updatedData.Subject
	announcement.Date = updatedData.Date
	announcement.Type = updatedData.Type

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
