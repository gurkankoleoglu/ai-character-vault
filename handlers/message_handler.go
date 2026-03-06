package handlers

import (
	"net/http"
	"stajyer-backend/config"
	"stajyer-backend/models"

	"github.com/gin-gonic/gin"
)

// SendMessage: Bir kullanıcıdan diğerine DM gönderir
func SendMessage(c *gin.Context) {
	var req struct {
		SenderPhone   string `json:"sender_phone" binding:"required"`
		ReceiverPhone string `json:"receiver_phone" binding:"required"`
		Content       string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Eksik bilgi formatı veya boş mesaj"})
		return
	}

	if req.SenderPhone == req.ReceiverPhone {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kendinize mesaj gönderemezsiniz"})
		return
	}

	var sender, receiver models.User
	if err := config.DB.Where("phone = ?", req.SenderPhone).First(&sender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Gönderen bulunamadı"})
		return
	}
	if err := config.DB.Where("phone = ?", req.ReceiverPhone).First(&receiver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Alıcı bulunamadı"})
		return
	}

	// Mesajı oluştur ve kaydet
	msg := models.Message{
		SenderID:   sender.ID,
		ReceiverID: receiver.ID,
		Content:    req.Content,
		IsRead:     false,
	}

	if err := config.DB.Create(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Mesaj gönderilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Mesaj gönderildi", "data": msg})
}

// GetConversation: İki kişi arasındaki sohbet geçmişini tarihe göre sıralı (eskiden yeniye) getirir
func GetConversation(c *gin.Context) {
	// URL'den örneğin /messages/5551112233/5559998877 şeklinde gelecek
	phone1 := c.Param("phone1")
	phone2 := c.Param("phone2")

	var user1, user2 models.User
	if err := config.DB.Where("phone = ?", phone1).First(&user1).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "1. Kullanıcı bulunamadı"})
		return
	}
	if err := config.DB.Where("phone = ?", phone2).First(&user2).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "2. Kullanıcı bulunamadı"})
		return
	}

	var messages []models.Message

	// (Gönderen=U1 ve Alıcı=U2) VEYA (Gönderen=U2 ve Alıcı=U1)
	// Order("created_at asc") -> Eskiden yeniye sıralar ki frontend'de chat balonları doğru dizilsin
	config.DB.Preload("Sender").Preload("Receiver").
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			user1.ID, user2.ID, user2.ID, user1.ID).
		Order("created_at asc").
		Find(&messages)

	// Sohbet açıldığında okunmamış mesajları "okundu" (IsRead = true) olarak işaretleyebiliriz (İsteğe bağlı)
	config.DB.Model(&models.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = ?", user2.ID, user1.ID, false).
		Update("is_read", true)

	c.JSON(http.StatusOK, messages)
}
