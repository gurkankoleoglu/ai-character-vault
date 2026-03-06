package models

import "time"

// Message: İki kullanıcı arasındaki özel mesajları (DM) tutar
type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SenderID   uint      `gorm:"not null" json:"sender_id"`
	Sender     User      `gorm:"foreignKey:SenderID" json:"sender"` // Gönderen
	ReceiverID uint      `gorm:"not null" json:"receiver_id"`
	Receiver   User      `gorm:"foreignKey:ReceiverID" json:"receiver"` // Alıcı
	Content    string    `gorm:"type:text;not null" json:"content"`     // Mesaj içeriği
	IsRead     bool      `gorm:"default:false" json:"is_read"`          // Okundu bilgisi
	CreatedAt  time.Time `json:"created_at"`
}
