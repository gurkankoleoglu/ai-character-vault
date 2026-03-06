package models

import "time"

// Connection: Kullanıcılar arası bağlantı/arkadaşlık isteklerini tutar
type Connection struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	SenderID   uint      `gorm:"not null" json:"sender_id"`
	Sender     User      `gorm:"foreignKey:SenderID" json:"sender"` // İsteği atan kişi
	ReceiverID uint      `gorm:"not null" json:"receiver_id"`
	Receiver   User      `gorm:"foreignKey:ReceiverID" json:"receiver"` // İsteği alan kişi
	Status     string    `gorm:"default:'beklemede'" json:"status"`     // 'beklemede', 'onaylandi', 'reddedildi'
	CreatedAt  time.Time `json:"created_at"`
}
