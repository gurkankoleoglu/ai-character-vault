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
		CompanyPhone string `json:"company_phone" binding:"required"`
		Title        string `json:"title" binding:"required"`
		Quota        int    `json:"quota"`
		Location     string `json:"location"`
		Type         string `json:"type"`     // Frontend'in gönderdiği çalışma tipi
		Category     string `json:"category"` // Frontend'in gönderdiği kategori (staj/iş)
		Details      string `json:"details"`
		Requirements string `json:"requirements"`
		Benefits     string `json:"benefits"` // Frontend'in gönderdiği imkanlar
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Eksik veya hatalı veri formatı"})
		return
	}

	var user models.User
	if err := config.DB.Where("phone = ? AND role = ?", req.CompanyPhone, "company").Preload("CompanyProfile").First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Yetkisiz işlem: Şirket hesabı bulunamadı"})
		return
	}

	if user.CompanyProfile == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lütfen önce şirket profilinizi tamamlayın"})
		return
	}

	// YENİ ALANLARLA BİRLİKTE VERİTABANINA KAYIT
	job := models.JobPosting{
		CompanyID:      user.CompanyProfile.ID,
		Title:          req.Title,
		Description:    req.Details,
		Requirements:   req.Requirements,
		Benefits:       req.Benefits, // Yeni eklendi
		WorkType:       req.Type,     // Yeni eklendi
		InternshipType: req.Category,
		Capacity:       req.Quota,
		Location:       req.Location,
		IsActive:       true,
	}

	if err := config.DB.Create(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İlan veritabanına kaydedilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "İlan başarıyla yayınlandı!", "job": job})
}

// GetAllJobs: Tüm aktif ilanları listeler
func GetAllJobs(c *gin.Context) {
	var jobs []models.JobPosting

	if err := config.DB.Preload("Company").Where("is_active = ?", true).Order("created_at desc").Find(&jobs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İlanlar çekilirken bir hata oluştu"})
		return
	}

	var response []gin.H
	for _, job := range jobs {
		var user models.User
		config.DB.Where("id = ?", job.Company.UserID).First(&user)

		response = append(response, gin.H{
			"id":              job.ID,
			"title":           job.Title,
			"description":     job.Description,
			"requirements":    job.Requirements,
			"benefits":        job.Benefits, // Yeni eklendi
			"work_type":       job.WorkType, // Yeni eklendi
			"internship_type": job.InternshipType,
			"capacity":        job.Capacity,
			"location":        job.Location,
			"created_at":      job.CreatedAt,
			"company_name":    job.Company.CompanyName,
			"company_phone":   user.Phone,
		})
	}

	if response == nil {
		response = make([]gin.H, 0)
	}

	c.JSON(http.StatusOK, response)
}
