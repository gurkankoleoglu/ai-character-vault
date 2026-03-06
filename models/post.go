package models

import "time"

// Post: Kullanıcıların (Öğrenci veya Şirket) anasayfada paylaştığı gönderiler
type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID" json:"user"` // Gönderiyi kim paylaştı?
	Content   string    `gorm:"type:text" json:"content"`      // Gönderi metni
	ImageURL  string    `json:"image_url"`                     // Varsa yüklenen resmin linki
	Likes     int       `gorm:"default:0" json:"likes"`        // Beğeni sayısı
	CreatedAt time.Time `json:"created_at"`
}
