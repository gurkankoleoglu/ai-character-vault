package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"stajyer-backend/config"
	"stajyer-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreatePost: Yeni bir gönderi paylaşır
func CreatePost(c *gin.Context) {
	phone := c.PostForm("phone")
	content := c.PostForm("content")

	if content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gönderi içeriği boş olamaz"})
		return
	}

	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	imageFile, _ := c.FormFile("image_file")
	var imagePath string

	if imageFile != nil {
		imageUUID := uuid.New().String()
		imagePath = filepath.ToSlash(filepath.Join("uploads", fmt.Sprintf("post_%s%s", imageUUID, filepath.Ext(imageFile.Filename))))
		if err := c.SaveUploadedFile(imageFile, imagePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Resim yüklenirken bir hata oluştu"})
			return
		}
	}

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

// GetFeed: Anasayfadaki tüm gönderileri (Yorumlar ve Beğenenler ile) getirir
func GetFeed(c *gin.Context) {
	var posts []models.Post

	// Preload zinciri: Post'u atan kişi + Yorumlar + Yorumu atan kişiler + Beğeniler
	if err := config.DB.
		Preload("User").
		Preload("Comments.User"). // Yorumları atan kullanıcıların bilgileri
		Preload("PostLikes").     // Postu kimlerin beğendiği
		Order("created_at desc").
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Akış yüklenemedi"})
		return
	}

	c.JSON(http.StatusOK, posts)
}

// ToggleLike: Gönderiyi beğenir, zaten beğenilmişse beğeniyi geri çeker
func ToggleLike(c *gin.Context) {
	postID := c.Param("id")
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefon numarası gerekli"})
		return
	}

	var user models.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	var post models.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gönderi bulunamadı"})
		return
	}

	// Bu kullanıcı bu gönderiyi daha önce beğenmiş mi?
	var existingLike models.PostLike
	err := config.DB.Where("post_id = ? AND user_id = ?", post.ID, user.ID).First(&existingLike).Error

	if err == gorm.ErrRecordNotFound {
		// Beğenilmemişse BEĞEN
		like := models.PostLike{PostID: post.ID, UserID: user.ID}
		config.DB.Create(&like)
		config.DB.Model(&post).Update("likes", gorm.Expr("likes + ?", 1))

		c.JSON(http.StatusOK, gin.H{"message": "Gönderi beğenildi", "action": "liked"})
	} else {
		// Zaten beğenilmişse BEĞENİYİ GERİ ÇEK
		config.DB.Delete(&existingLike)
		config.DB.Model(&post).Update("likes", gorm.Expr("likes - ?", 1))

		c.JSON(http.StatusOK, gin.H{"message": "Beğeni geri çekildi", "action": "unliked"})
	}
}

// AddComment: Gönderiye yorum yapar
func AddComment(c *gin.Context) {
	postID := c.Param("id")
	var req struct {
		Phone   string `json:"phone" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Eksik bilgi: telefon ve content gerekli"})
		return
	}

	var user models.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	var post models.Post
	if err := config.DB.First(&post, postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gönderi bulunamadı"})
		return
	}

	comment := models.Comment{
		PostID:  post.ID,
		UserID:  user.ID,
		Content: req.Content,
	}

	if err := config.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Yorum eklenemedi"})
		return
	}

	// Yorum eklendikten sonra, ön yüze gönderirken yazarın bilgilerini de bağlayıp yolluyoruz
	config.DB.Preload("User").First(&comment, comment.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Yorum başarıyla eklendi", "comment": comment})
}
