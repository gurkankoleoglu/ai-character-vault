package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"stajyer-backend/config"
	"stajyer-backend/models"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateApplication: Yeni başvuru oluşturur ve ilana bağlar
func CreateApplication(c *gin.Context) {
	phone := c.PostForm("phone")
	adSoyad := c.PostForm("full_name")
	email := c.PostForm("email")
	okul := c.PostForm("university")
	bolum := c.PostForm("department")
	stajTuru := c.PostForm("intern_type")
	stajTarihleri := c.PostForm("stajTarihleri")
	jobPostingID := c.PostForm("job_posting_id")

	if jobPostingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lütfen başvurmak istediğiniz ilanı belirtin."})
		return
	}

	var existingApp models.Application
	if err := config.DB.Where("phone = ? AND job_posting_id = ? AND durum IN ?", phone, jobPostingID, []string{"Beklemede", "Onaylandı"}).First(&existingApp).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bu ilana zaten devam eden veya onaylanmış bir başvurunuz var."})
		return
	}

	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err == nil {
		config.DB.Model(&user).Updates(models.User{
			FullNames:          adSoyad,
			Email:              email,
			IsProfileCompleted: true,
		})

		var student models.Student
		if err := config.DB.Where("user_id = ?", user.ID).First(&student).Error; err != nil {
			student = models.Student{
				UserID:     user.ID,
				University: okul,
				Department: bolum,
			}
			config.DB.Create(&student)
		} else {
			config.DB.Model(&student).Updates(models.Student{
				University: okul,
				Department: bolum,
			})
		}
	}

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

	var job models.JobPosting
	if err := config.DB.First(&job, jobPostingID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Başvurulmak istenen ilan bulunamadı veya yayından kaldırılmış."})
		return
	}
	app.JobPostingID = job.ID

	if err := config.DB.Create(&app).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Veritabanı hatası"})
		return
	}

	config.DB.Model(&job).UpdateColumn("application_count", gorm.Expr("application_count + ?", 1))

	c.JSON(http.StatusOK, gin.H{"message": "Başvurunuz alındı, şirket onayı bekleniyor.", "status": "Beklemede"})
}

// UpdateApplicationStatus: Admin Onay/Red yaptığında çalışır ve KONSOLA SMS BASAR
func UpdateApplicationStatus(c *gin.Context) {
	id := c.Param("id")
	status := c.Query("status")

	var application models.Application
	if err := config.DB.First(&application, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Başvuru bulunamadı"})
		return
	}

	if err := config.DB.Model(&application).Update("durum", status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Durum güncellenemedi"})
		return
	}

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

// GetApplications: Tüm başvuruları listeler (Frontend'in beklediği formatta)
func GetApplications(c *gin.Context) {
	var apps []models.Application
	config.DB.Order("created_at desc").Find(&apps)

	var response []gin.H
	for _, app := range apps {
		// Başvurunun yapıldığı ilanı bul
		var job models.JobPosting
		config.DB.Where("id = ?", app.JobPostingID).First(&job)

		// İlanın sahibini (Şirket) bul
		var company models.Company
		config.DB.Where("id = ?", job.CompanyID).First(&company)

		// Şirket yetkilisini (User) bul (Telefon numarası için)
		var user models.User
		config.DB.Where("id = ?", company.UserID).First(&user)

		// Durumu İngilizce etiketlere çevir (Frontend böyle bekliyor)
		status := "pending"
		if app.Durum == "Onaylandı" {
			status = "accepted"
		} else if app.Durum == "Reddedildi" {
			status = "rejected"
		}

		response = append(response, gin.H{
			"id":             app.ID,
			"job_posting_id": app.JobPostingID,
			"job_title":      job.Title,
			"student_name":   app.FullNames,
			"student_email":  app.Email,
			"status":         status,
			"cv_url":         "http://localhost:8080/" + app.CVPath,
			"letter_url":     "http://localhost:8080/" + app.LetterPath,
			"company_phone":  user.Phone, // Frontend filtresi için çok önemli
		})
	}

	if response == nil {
		response = make([]gin.H, 0)
	}

	c.JSON(http.StatusOK, response)
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
	c.JSON(http.StatusOK, gin.H{
		"has_application": true,
		"status":          app.Durum,
		"application":     app,
	})
}
