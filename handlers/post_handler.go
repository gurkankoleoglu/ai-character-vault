package handlers

import (
	"net/http"

	"stajyer-backend/config"
	"stajyer-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Beğeni kontrolü sırasında hata oluştu"})
		return
	}

	if err == gorm.ErrRecordNotFound {
		// Beğenilmemişse BEĞEN
		like := models.PostLike{PostID: post.ID, UserID: user.ID}
		config.DB.Create(&like)
		config.DB.Model(&post).Update("likes", gorm.Expr("likes + ?", 1))

		c.JSON(http.StatusOK, gin.H{"message": "Gönderi beğenildi", "action": "liked"})
		return
	}

	// Zaten beğenilmişse BEĞENİYİ GERİ ÇEK
	config.DB.Delete(&existingLike)
	config.DB.Model(&post).Update("likes", gorm.Expr("CASE WHEN likes > 0 THEN likes - 1 ELSE 0 END"))

	c.JSON(http.StatusOK, gin.H{"message": "Beğeni geri çekildi", "action": "unliked"})
}

// AddComment: Gönderiye yorum yapar
func AddComment(c *gin.Context) {
	// (kalan orijinal kod burada)
}
