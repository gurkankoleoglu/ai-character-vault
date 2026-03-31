package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"

	"stajyer-backend/config"
	"stajyer-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreatePost: Yeni bir gönderi oluşturur (Resim ve Görünürlük desteği ile)
func CreatePost(c *gin.Context) {
	phone := c.PostForm("phone")
	content := c.PostForm("content")
	visibility := c.PostForm("visibility")

	if phone == "" || content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefon numarası ve içerik gerekli"})
		return
	}

	// Görünürlük seçilmemişse varsayılan olarak "public" (herkese açık) yap
	if visibility == "" {
		visibility = "public"
	}

	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	// Resim yükleme işlemi
	var imageURL string
	file, err := c.FormFile("image_file")
	if err == nil {
		// Çakışmaları önlemek için dosya isminin başına benzersiz bir zaman damgası ekliyoruz
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		filePath := "uploads/" + filename

		// Dosyayı sunucuya (uploads klasörüne) kaydet
		if err := c.SaveUploadedFile(file, filePath); err == nil {
			imageURL = filePath
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Resim kaydedilirken bir sorun oluştu"})
			return
		}
	}

	// Yeni post nesnesini oluştur
	newPost := models.Post{
		UserID:     user.ID,
		Content:    content,
		Visibility: visibility,
		ImageURL:   imageURL,
	}

	if err := config.DB.Create(&newPost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gönderi paylaşılamadı"})
		return
	}

	// Yeni gönderiyi kullanıcı bilgileriyle birlikte frontend'e döndür
	if err := config.DB.Preload("User").First(&newPost, newPost.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gönderi yüklendi ama geri döndürülemedi"})
		return
	}

	c.JSON(http.StatusCreated, newPost)
}

// GetFeed: Anasayfadaki tüm gönderileri (Yorumlar ve Beğenenler ile) getirir
func GetFeed(c *gin.Context) {
	var posts []models.Post

	if err := config.DB.
		Preload("User").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at desc")
		}).
		Preload("Comments.User").
		Preload("PostLikes").
		Order("created_at desc").
		Find(&posts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Akış yüklenemedi"})
		return
	}

	// Her post'un likes sayısını post_likes tablosundan senkronize et
	for i := range posts {
		posts[i].Likes = (len(posts[i].PostLikes))
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

	var existingLike models.PostLike
	err := config.DB.Where("post_id = ? AND user_id = ?", post.ID, user.ID).First(&existingLike).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Beğeni kontrolü sırasında hata oluştu"})
		return
	}

	if err == gorm.ErrRecordNotFound {
		// Beğenilmemişse BEĞEN
		like := models.PostLike{PostID: post.ID, UserID: user.ID}
		if err := config.DB.Create(&like).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Beğeni eklenemedi"})
			return
		}
		config.DB.Model(&post).UpdateColumn("likes", gorm.Expr("likes + 1"))

		// Güncel beğeni sayısını döndür
		var likeCount int64
		config.DB.Model(&models.PostLike{}).Where("post_id = ?", post.ID).Count(&likeCount)

		c.JSON(http.StatusOK, gin.H{
			"message": "Gönderi beğenildi",
			"action":  "liked",
			"likes":   likeCount,
		})
		return
	}

	// Zaten beğenilmişse BEĞENİYİ GERİ ÇEK
	if err := config.DB.Delete(&existingLike).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Beğeni kaldırılamadı"})
		return
	}
	config.DB.Model(&post).UpdateColumn("likes", gorm.Expr("CASE WHEN likes > 0 THEN likes - 1 ELSE 0 END"))

	// Güncel beğeni sayısını döndür
	var likeCount int64
	config.DB.Model(&models.PostLike{}).Where("post_id = ?", post.ID).Count(&likeCount)

	c.JSON(http.StatusOK, gin.H{
		"message": "Beğeni geri çekildi",
		"action":  "unliked",
		"likes":   likeCount,
	})
}

// AddComment: Gönderiye yorum ekler ve yorum sahibinin bilgileriyle birlikte döndürür
func AddComment(c *gin.Context) {
	postID := c.Param("id")

	var req struct {
		Phone   string `json:"phone" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefon ve yorum içeriği gerekli"})
		return
	}

	if len([]rune(req.Content)) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Yorum en fazla 500 karakter olabilir"})
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

	// Yorumu kullanıcı bilgileriyle birlikte yeniden yükle
	if err := config.DB.Preload("User").First(&comment, comment.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Yorum yüklenemedi"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Yorum eklendi",
		"comment": comment,
	})
}

// DeleteComment: Kullanıcının kendi yorumunu silmesini sağlar
func DeleteComment(c *gin.Context) {
	// URL'den silinecek yorumun ID'sini alıyoruz (Örn: /comments/:id)
	commentID := c.Param("id")

	var req struct {
		Phone string `json:"phone" binding:"required"`
	}

	// İstek gövdesinden telefon numarasını al
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Telefon numarası gerekli"})
		return
	}

	// Kullanıcıyı telefon numarasından bul
	var user models.User
	if err := config.DB.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	// Silinecek yorumu veritabanından bul
	var comment models.Comment
	if err := config.DB.First(&comment, commentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Yorum bulunamadı"})
		return
	}

	// Güvenlik Kontrolü: Yorum bu kullanıcıya mı ait?
	if comment.UserID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bu yorumu silme yetkiniz yok"})
		return
	}

	// Yorumu sil
	if err := config.DB.Delete(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Yorum silinemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Yorum başarıyla silindi",
		"action":  "deleted",
	})
}
