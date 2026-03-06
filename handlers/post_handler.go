package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"stajyer-backend/config"
	"stajyer-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreatePost: Yeni bir gönderi (metin + opsiyonel resim) paylaşır
func CreatePost(c *gin.Context) {
	phone := c.PostForm("phone")     // Kim paylaşıyor?
	content := c.PostForm("content") // Ne paylaşıyor?

	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gönderi içeriği boş olamaz"})
		return
	}

	// Kullanıcıyı bul
	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	// İsteğe bağlı: Resim dosyası var mı kontrol et
	imageFile, _ := c.FormFile("image_file")
	var imagePath string

	if imageFile != nil {
		// Resmi uploads klasörüne benzersiz bir isimle kaydet
		imageUUID := uuid.New().String()
		imagePath = filepath.ToSlash(filepath.Join("uploads", fmt.Sprintf("post_%s%s", imageUUID, filepath.Ext(imageFile.Filename))))
		if err := c.SaveUploadedFile(imageFile, imagePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Resim yüklenirken bir hata oluştu"})
			return
		}
	}

	// Gönderiyi veritabanına kaydet
	post := models.Post{
		UserID:   user.ID,
		Content:  content,
		ImageURL: imagePath,
	}

	if err := config.DB.Create(&post).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gönderi paylaşılamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Gönderiniz başarıyla paylaşıldı!", "post": post})
}

// GetFeed: Anasayfadaki tüm gönderileri (paylaşan kişinin bilgileriyle) en yeniler en üstte olacak şekilde getirir
func GetFeed(c *gin.Context) {
	var posts []models.Post

	// Preload ile gönderiyi paylaşan "User" bilgisini de JSON'a dahil ediyoruz
	if err := config.DB.Preload("User").Order("created_at desc").Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Akış yüklenemedi"})
		return
	}

	c.JSON(http.StatusOK, posts)
}
