package handlers

import (
	"net/http"
	"stajyer-backend/config"
	"stajyer-backend/models"
	"time"

	"github.com/gin-gonic/gin"
)

func GetInternProgress(c *gin.Context) {
	phone := c.Param("phone")
	var user models.User

	// Kullanıcıyı bul
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	// Eğer tarihler henüz girilmemişse hata verme, 0 döndür
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

	// Kalan gün (Bitiş - Bugün)
	remaining := int(user.InternEndDate.Sub(now).Hours() / 24)
	if remaining < 0 {
		remaining = 0
	}

	// Toplam süre ve geçen süre
	totalDuration := user.InternEndDate.Sub(*user.InternStartDate).Hours() / 24
	elapsed := now.Sub(*user.InternStartDate).Hours() / 24

	if elapsed < 0 {
		elapsed = 0
	} // Staj henüz başlamadıysa

	// Yüzde hesaplama (Bölme hatasını önlemek için kontrol)
	percent := 0.0
	if totalDuration > 0 {
		percent = (elapsed / totalDuration) * 100
	}
	if percent > 100 {
		percent = 100
	}

	c.JSON(http.StatusOK, gin.H{
		"remaining_days": remaining,
		"elapsed_days":   int(elapsed),
		"percentage":     int(percent),
	})
}

// SaveDailyReport: Kullanıcının yazdığı raporu veritabanına kaydeder
func SaveDailyReport(c *gin.Context) {
	var req struct {
		Phone  string    `json:"phone"`
		Report string    `json:"report"`
		Date   time.Time `json:"date"` // Takvimden gelen tarih
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

	// Models kısmında oluşturduğumuz DailyReport yapısına uygun kayıt
	newReport := models.DailyReport{
		UserID:  user.ID,
		Content: req.Report, // Modelindeki alan Content ise Content yap
		Date:    req.Date,
	}

	if err := config.DB.Create(&newReport).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Rapor kaydedilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rapor başarıyla kaydedildi!"})
}

// GetUserReports: Kullanıcının tüm geçmiş raporlarını getirir
func GetUserReports(c *gin.Context) {
	phone := c.Param("phone")
	var user models.User

	// Önce kullanıcıyı bulalım
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	var reports []models.DailyReport
	// Kullanıcıya ait raporları tarihe göre yeniden eskiye (desc) sıralayıp getir
	if err := config.DB.Where("user_id = ?", user.ID).Order("date DESC").Find(&reports).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Raporlar getirilemedi"})
		return
	}

	c.JSON(http.StatusOK, reports)
}

// UpdateReport: Veritabanındaki mevcut bir raporu günceller
func UpdateReport(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri"})
		return
	}

	// ID'ye göre raporu bul ve içeriğini güncelle
	if err := config.DB.Model(&models.DailyReport{}).Where("id = ?", id).Update("content", req.Content).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Güncelleme başarısız"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rapor güncellendi"})
}

// DeleteReport: Bir raporu veritabanından kalıcı olarak siler
func DeleteReport(c *gin.Context) {
	id := c.Param("id")

	if err := config.DB.Delete(&models.DailyReport{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Silme işlemi başarısız"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rapor başarıyla silindi"})
}
