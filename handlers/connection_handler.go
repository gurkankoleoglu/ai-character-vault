package handlers

import (
	"net/http"
	"stajyer-backend/config"
	"stajyer-backend/models"

	"github.com/gin-gonic/gin"
)

// SendConnectionRequest: Arkadaşlık/Bağlantı isteği gönderir
func SendConnectionRequest(c *gin.Context) {
	var req struct {
		SenderPhone   string `json:"sender_phone" binding:"required"`
		ReceiverPhone string `json:"receiver_phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Eksik bilgi formatı"})
		return
	}

	if req.SenderPhone == req.ReceiverPhone {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kendinize bağlantı isteği gönderemezsiniz"})
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

	// Daha önce istek atılmış mı kontrolü
	var existingConn models.Connection
	err := config.DB.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		sender.ID, receiver.ID, receiver.ID, sender.ID).First(&existingConn).Error

	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Bu kullanıcıyla zaten bir bağlantınız veya bekleyen isteğiniz var",
			"status": existingConn.Status,
		})
		return
	}

	conn := models.Connection{
		SenderID:   sender.ID,
		ReceiverID: receiver.ID,
		Status:     "beklemede",
	}

	if err := config.DB.Create(&conn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "İstek gönderilemedi"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bağlantı isteği gönderildi!", "connection": conn})
}

// RespondConnectionRequest: İsteği onaylar veya reddeder
func RespondConnectionRequest(c *gin.Context) {
	connID := c.Param("id")
	status := c.Query("action") // "onaylandi" veya "reddedildi"

	var conn models.Connection
	if err := config.DB.First(&conn, connID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bağlantı isteği bulunamadı"})
		return
	}

	conn.Status = status
	if err := config.DB.Save(&conn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Güncelleme başarısız"})
		return
	}

	message := "Bağlantı isteği reddedildi."
	if status == "onaylandi" {
		message = "Artık bağlantı kurdunuz!"
	}

	c.JSON(http.StatusOK, gin.H{"message": message, "status": status})
}

// GetConnectionStatus: İki kullanıcı arasındaki bağlantı durumunu kontrol eder (Profil sayfası için)
func GetConnectionStatus(c *gin.Context) {
	p1 := c.Param("phone1")
	p2 := c.Param("phone2")

	var user1, user2 models.User
	// Kullanıcıların varlığını ve ID'lerini al
	if err := config.DB.Where("phone = ?", p1).First(&user1).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "none"})
		return
	}
	if err := config.DB.Where("phone = ?", p2).First(&user2).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "none"})
		return
	}

	var conn models.Connection
	// Her iki yönde de kayıt var mı bak (A -> B veya B -> A)
	err := config.DB.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		user1.ID, user2.ID, user2.ID, user1.ID).First(&conn).Error

	if err != nil {
		// Bağlantı kaydı yoksa
		c.JSON(http.StatusOK, gin.H{"status": "none"})
		return
	}

	// Kayıt varsa veritabanındaki güncel durumu dön
	c.JSON(http.StatusOK, gin.H{"status": conn.Status})
}

// GetPendingRequests: Kullanıcıya gelen (bekleyen) istekleri listeler
func GetPendingRequests(c *gin.Context) {
	phone := c.Param("phone")

	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	var requests []models.Connection
	config.DB.Preload("Sender").
		Where("receiver_id = ? AND status = ?", user.ID, "beklemede").
		Find(&requests)

	c.JSON(http.StatusOK, requests)
}

// GetConnections: Kullanıcının onaylanmış tüm bağlantılarını listeler
func GetConnections(c *gin.Context) {
	phone := c.Param("phone")
	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	var connections []models.Connection
	config.DB.Preload("Sender").Preload("Receiver").
		Where("status = ? AND (sender_id = ? OR receiver_id = ?)", "onaylandi", user.ID, user.ID).
		Find(&connections)

	var friends []models.User
	for _, conn := range connections {
		if conn.SenderID == user.ID {
			friends = append(friends, conn.Receiver)
		} else {
			friends = append(friends, conn.Sender)
		}
	}

	c.JSON(http.StatusOK, friends)
}

// RemoveConnection: Var olan bir bağlantıyı (veya isteği) tamamen siler
func RemoveConnection(c *gin.Context) {
	phone1 := c.Param("phone1")
	phone2 := c.Param("phone2")

	var user1, user2 models.User
	if err := config.DB.Where("phone = ?", phone1).First(&user1).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}
	if err := config.DB.Where("phone = ?", phone2).First(&user2).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Karşı kullanıcı bulunamadı"})
		return
	}

	// İki kullanıcı arasındaki bağlantı kaydını kalıcı olarak sil
	if err := config.DB.Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
		user1.ID, user2.ID, user2.ID, user1.ID).Delete(&models.Connection{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Bağlantı kaldırılamadı"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Bağlantı başarıyla silindi"})
}
