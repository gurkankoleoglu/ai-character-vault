package handlers

import (
	"net/http"
	"stajyer-backend/config"
	"stajyer-backend/models"
	"time"

	"github.com/gin-gonic/gin"
)

// GetInternProgress: Kullanıcının staj tarihlerine göre kalan gün ve yüzde hesaplar
func GetInternProgress(c *gin.Context) {
	phone := c.Param("phone")
	var user models.User

	// Kullanıcıyı veritabanından bul
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	// Eğer staj tarihleri henüz girilmemişse (nil ise) varsayılan değerler dön
	if user.InternStartDate == nil || user.InternEndDate == nil {
		c.JSON(http.StatusOK, gin.H{
			"remaining_days": 0,
			"elapsed_days":   0,
			"percentage":     0,
			"message":        "Staj tarihleri henüz tanımlanmamış",
		})
		return
	}

	now := time.Now()

	// Kalan gün hesabı (Bitiş Tarihi - Bugün)
	remaining := int(user.InternEndDate.Sub(now).Hours() / 24)
	if remaining < 0 {
		remaining = 0
	}

	// Toplam staj süresi ve geçen süre (Gün bazında)
	totalDuration := user.InternEndDate.Sub(*user.InternStartDate).Hours() / 24
	elapsed := now.Sub(*user.InternStartDate).Hours() / 24

	if elapsed < 0 {
		elapsed = 0
	}

	// Yüzde hesaplama (0'a bölünme hatasını önlemek için)
	percent := 0.0
	if totalDuration > 0 {
		percent = (elapsed / totalDuration) * 100
	}

	// Sınırlandırmalar (%0 - %100 arası)
	if percent > 100 {
		percent = 100
	}
	if percent < 0 {
		percent = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"remaining_days": remaining,
		"elapsed_days":   int(elapsed),
		"percentage":     int(percent),
	})
}

// SaveDailyReport: Belirli bir tarih için rapor kaydeder veya varsa günceller (Upsert)
func SaveDailyReport(c *gin.Context) {
	var req struct {
		Phone  string    `json:"phone" binding:"required"`
		Report string    `json:"report" binding:"required"`
		Date   time.Time `json:"date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri formatı"})
		return
	}

	var user models.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	// Aynı güne ait mükerrer kaydı önlemek için kontrol
	var existingReport models.DailyReport
	// PostgreSQL'de DATE() fonksiyonu ile sadece gün/ay/yıl bazlı kontrol yapıyoruz
	err := config.DB.Where("user_id = ? AND DATE(date) = DATE(?)", user.ID, req.Date).First(&existingReport).Error

	if err == nil {
		// Eğer o güne ait rapor varsa, içeriği güncelle (Update)
		existingReport.Content = req.Report
		config.DB.Save(&existingReport)
		c.JSON(http.StatusOK, gin.H{"message": "Mevcut rapor güncellendi!"})
	} else {
		// Eğer o güne ait rapor yoksa, yeni kayıt oluştur (Create)
		newReport := models.DailyReport{
			UserID:  user.ID,
			Content: req.Report,
			Date:    req.Date,
		}
		if err := config.DB.Create(&newReport).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rapor kaydedilemedi"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Yeni rapor başarıyla kaydedildi!"})
	}
}

// GetUserReports: Kullanıcının tüm raporlarını tarihe göre sıralı getirir
func GetUserReports(c *gin.Context) {
	phone := c.Param("phone")
	var user models.User

	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	var reports []models.DailyReport
	// Yeniden eskiye sıralama (En son yazılan en üstte)
	if err := config.DB.Where("user_id = ?", user.ID).Order("date DESC").Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Raporlar çekilemedi"})
		return
	}

	c.JSON(http.StatusOK, reports)
}

// UpdateReport: Raporlarım sayfasından gelen manuel düzenleme isteği
func UpdateReport(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "İçerik boş olamaz"})
		return
	}

	if err := config.DB.Model(&models.DailyReport{}).Where("id = ?", id).Update("content", req.Content).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Rapor güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rapor başarıyla güncellendi"})
}

// DeleteReport: Belirli bir raporu kalıcı olarak siler
func DeleteReport(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Delete(&models.DailyReport{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Silme işlemi başarısız"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rapor başarıyla silindi"})
}
