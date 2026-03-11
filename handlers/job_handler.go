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
		Type         string `json:"type"`
		Category     string `json:"category"`
		Details      string `json:"details"`
		Requirements string `json:"requirements"`
		Benefits     string `json:"benefits"`
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

	job := models.JobPosting{
		CompanyID:      user.CompanyProfile.ID,
		Title:          req.Title,
		Description:    req.Details,
		Requirements:   req.Requirements,
		Benefits:       req.Benefits,
		WorkType:       req.Type,
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
			"benefits":        job.Benefits,
			"work_type":       job.WorkType,
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

// YENİ: İlan Güncelleme (Update)
func UpdateJob(c *gin.Context) {
	jobID := c.Param("id")
	var req struct {
		Title        string `json:"title" binding:"required"`
		Quota        int    `json:"quota"`
		Location     string `json:"location"`
		Type         string `json:"type"`
		Category     string `json:"category"`
		Details      string `json:"details"`
		Requirements string `json:"requirements"`
		Benefits     string `json:"benefits"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz veri formatı"})
		return
	}

	var job models.JobPosting
	if err := config.DB.First(&job, jobID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "İlan bulunamadı"})
		return
	}

	job.Title = req.Title
	job.Description = req.Details
	job.Requirements = req.Requirements
	job.Benefits = req.Benefits
	job.WorkType = req.Type
	job.InternshipType = req.Category
	job.Capacity = req.Quota
	job.Location = req.Location

	if err := config.DB.Save(&job).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İlan güncellenemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "İlan başarıyla güncellendi"})
}

// YENİ: İlan Silme (Delete) - KESİN VE FİZİKSEL SİLME (Unscoped)
func DeleteJob(c *gin.Context) {
	jobID := c.Param("id")

	// 1. ADIM: Başvuruları KALICI OLARAK veritabanından kazı (Unscoped)
	if err := config.DB.Unscoped().Where("job_posting_id = ?", jobID).Delete(&models.Application{}).Error; err != nil {
		// Hata olsa bile devam et (Belki başvuru hiç yoktur)
	}

	// 2. ADIM: İlanın kendisini KALICI OLARAK sil (Unscoped)
	if err := config.DB.Unscoped().Delete(&models.JobPosting{}, jobID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İlan silinemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "İlan ve bağlı başvurular veritabanından tamamen silindi"})
}
