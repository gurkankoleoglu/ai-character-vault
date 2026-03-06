package handlers

import (
	"net/http"
	"stajyer-backend/config"
	"stajyer-backend/models"

	"github.com/gin-gonic/gin"
)

// GetPendingApprovals: Onay bekleyen öğrencileri ve şirketleri listeler
// GetPendingApprovals: Onay bekleyen öğrencileri ve şirketleri listeler
func GetPendingApprovals(c *gin.Context) {
	// JSON'da "null" yerine "[]" (boş liste) görünmesi için önceden tanımlıyoruz
	pendingStudents := []models.Student{}
	pendingCompanies := []models.Company{}

	// Joins("User") kısımlarını kaldırdık çünkü modellerde back-relation tanımlamadık.
	// Bu sayede sorgular sorunsuz çalışacak.
	config.DB.Where("is_cert_approved = ?", false).Find(&pendingStudents)
	config.DB.Where("is_approved = ?", false).Find(&pendingCompanies)

	c.JSON(http.StatusOK, gin.H{
		"pending_students":  pendingStudents,
		"pending_companies": pendingCompanies,
	})
}

// ApproveStudent: Öğrenci belgesini onaylar veya reddeder
func ApproveStudent(c *gin.Context) {
	id := c.Param("id")         // Student tablosundaki ID
	action := c.Query("action") // URL'den ?action=approve gelecek

	var student models.Student
	if err := config.DB.First(&student, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Öğrenci profili bulunamadı"})
		return
	}

	if action == "approve" {
		student.IsCertApproved = true
		config.DB.Save(&student)
		c.JSON(http.StatusOK, gin.H{"message": "Öğrenci hesabı başarıyla onaylandı ve tam erişim verildi! ✅"})
	} else if action == "reject" {
		// Reddedilirse şimdilik sadece mesaj dönüyoruz (İleride silme eklenebilir)
		c.JSON(http.StatusOK, gin.H{"message": "Öğrenci belgesi reddedildi. ❌"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz işlem. Sadece 'approve' veya 'reject' kullanılabilir."})
	}
}

// ApproveCompany: Şirket hesabını onaylar veya reddeder
func ApproveCompany(c *gin.Context) {
	id := c.Param("id")         // Company tablosundaki ID
	action := c.Query("action") // URL'den ?action=approve gelecek

	var company models.Company
	if err := config.DB.First(&company, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Şirket profili bulunamadı"})
		return
	}

	if action == "approve" {
		company.IsApproved = true
		config.DB.Save(&company)
		c.JSON(http.StatusOK, gin.H{"message": "Şirket hesabı başarıyla onaylandı! Artık ilan açabilirler. ✅"})
	} else if action == "reject" {
		c.JSON(http.StatusOK, gin.H{"message": "Şirket hesabı reddedildi. ❌"})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz işlem."})
	}
}
