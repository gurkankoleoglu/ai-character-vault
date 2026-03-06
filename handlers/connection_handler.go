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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bu kullanıcıyla zaten bir bağlantınız veya bekleyen isteğiniz var", "status": existingConn.Status})
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
	connID := c.Param("id")     // Connection tablosundaki ID
	status := c.Query("action") // URL'den ?action=onaylandi gelecek

	if status != "onaylandi" && status != "reddedildi" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz işlem. Sadece 'onaylandi' veya 'reddedildi' olabilir."})
		return
	}

	var conn models.Connection
	if err := config.DB.First(&conn, connID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bağlantı isteği bulunamadı"})
		return
	}

	conn.Status = status
	config.DB.Save(&conn)

	c.JSON(http.StatusOK, gin.H{"message": "Bağlantı durumu güncellendi", "new_status": status})
}

// GetPendingRequests: Kullanıcıya gelen (bekleyen) istekleri listeler (Ağım sayfası)
func GetPendingRequests(c *gin.Context) {
	phone := c.Param("phone")

	var user models.User
	if err := config.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Kullanıcı bulunamadı"})
		return
	}

	var requests []models.Connection
	// Sadece Alıcısı(Receiver) olduğumuz ve durumu "beklemede" olanlar
	config.DB.Preload("Sender").
		Where("receiver_id = ? AND status = ?", user.ID, "beklemede").
		Find(&requests)

	c.JSON(http.StatusOK, requests)
}
