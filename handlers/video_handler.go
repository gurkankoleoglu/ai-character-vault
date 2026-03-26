package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"stajyer-backend/config"
	"stajyer-backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Ortak küçük fonksiyon: Resim yükleme işlemini yapar
func uploadThumbnail(c *gin.Context) string {
	file, _ := c.FormFile("thumbnail")
	if file != nil {
		ext := filepath.Ext(file.Filename)
		fileName := fmt.Sprintf("thumb_%s%s", uuid.New().String(), ext)
		filePath := filepath.Join("uploads", fileName)
		if err := c.SaveUploadedFile(file, filePath); err == nil {
			return "http://localhost:8080/uploads/" + fileName
		}
	}
	return ""
}

// ─── GET İŞLEMLERİ ──────────────────────────────────────────────

func GetPublicVideos(c *gin.Context) {
	var videos []models.Video
	config.DB.Where("is_company_only = ?", false).Order("date desc").Find(&videos)
	c.JSON(http.StatusOK, videos)
}

func GetCompanyVideos(c *gin.Context) {
	var videos []models.Video
	config.DB.Where("is_company_only = ?", true).Order("date desc").Find(&videos)
	c.JSON(http.StatusOK, videos)
}

// ─── CREATE İŞLEMLERİ ───────────────────────────────────────────

func createVideo(c *gin.Context, isCompanyOnly bool) {
	dateStr := c.PostForm("date")
	date, _ := time.Parse(time.RFC3339, dateStr)

	video := models.Video{
		Title:         c.PostForm("title"),
		Category:      c.PostForm("category"),
		Description:   c.PostForm("description"),
		VideoURL:      c.PostForm("video_url"),
		Duration:      c.PostForm("duration"),
		Date:          date,
		IsCompanyOnly: isCompanyOnly,
		ThumbnailURL:  uploadThumbnail(c),
	}

	if err := config.DB.Create(&video).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Video eklenemedi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": video})
}

func CreatePublicVideo(c *gin.Context)  { createVideo(c, false) }
func CreateCompanyVideo(c *gin.Context) { createVideo(c, true) }

// ─── UPDATE İŞLEMİ ──────────────────────────────────────────────

func UpdateVideo(c *gin.Context) {
	id := c.Param("id")
	var video models.Video

	if err := config.DB.First(&video, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Video bulunamadı"})
		return
	}

	video.Title = c.PostForm("title")
	video.Category = c.PostForm("category")
	video.Description = c.PostForm("description")
	video.VideoURL = c.PostForm("video_url")
	video.Duration = c.PostForm("duration")

	if dateStr := c.PostForm("date"); dateStr != "" {
		video.Date, _ = time.Parse(time.RFC3339, dateStr)
	}

	// Eğer yeni bir resim yüklendiyse eskisini ez
	if thumbUrl := uploadThumbnail(c); thumbUrl != "" {
		video.ThumbnailURL = thumbUrl
	}

	config.DB.Save(&video)
	c.JSON(http.StatusOK, gin.H{"data": video})
}

// ─── DELETE İŞLEMİ ──────────────────────────────────────────────

func DeleteVideo(c *gin.Context) {
	id := c.Param("id")
	config.DB.Delete(&models.Video{}, id)
	c.JSON(http.StatusOK, gin.H{"message": "Video silindi"})
}
