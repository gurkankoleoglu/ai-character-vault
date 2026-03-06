package handlers

import (
	"net/http"
	"stajyer-backend/config"
	"stajyer-backend/models"

	"github.com/gin-gonic/gin"
)

// CreateJob: Şirket yetkilisinin yeni bir ilan açması
func CreateJob(c *gin.Context) {
	var req struct {
		CompanyPhone   string `json:"company_phone" binding:"required"` // Hangi şirket açıyor?
		Title          string `json:"title" binding:"required"`
		Description    string `json:"description"`
		Requirements   string `json:"requirements"`
		InternshipType string `json:"internship_type"`
		Capacity       int    `json:"capacity"`
		Location       string `json:"location"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Eksik veya hatalı veri formatı"})
		return
	}

	// İlanı açan kişiyi ve şirket profilini bul
	var user models.User
	if err := config.DB.Where("phone = ? AND role = ?", req.CompanyPhone, "company").Preload("CompanyProfile").First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem: Şirket hesabı bulunamadı"})
		return
	}

	if user.CompanyProfile == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lütfen önce şirket profilinizi tamamlayın"})
		return
	}

	// İlanı oluştur
	job := models.JobPosting{
		CompanyID:      user.CompanyProfile.ID,
		Title:          req.Title,
		Description:    req.Description,
		Requirements:   req.Requirements,
		InternshipType: req.InternshipType,
		Capacity:       req.Capacity,
		Location:       req.Location,
		IsActive:       true,
	}

	if err := config.DB.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İlan veritabanına kaydedilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "İlan başarıyla yayınlandı!", "job": job})
}

// GetAllJobs: Sistemdeki tüm aktif ilanları (şirket bilgileriyle birlikte) listeler
func GetAllJobs(c *gin.Context) {
	var jobs []models.JobPosting

	// Preload ile ilanı açan şirketin bilgilerini de dahil ediyoruz
	if err := config.DB.Preload("Company").Where("is_active = ?", true).Order("created_at desc").Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İlanlar çekilirken bir hata oluştu"})
		return
	}

	c.JSON(http.StatusOK, jobs)
}
