package models

import "time"

// Post: Kullanıcıların (Öğrenci veya Şirket) anasayfada paylaştığı gönderiler
type Post struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null" json:"user_id"`
	User       User       `gorm:"foreignKey:UserID" json:"user"` // Gönderiyi kim paylaştı?
	Content    string     `gorm:"type:text" json:"content"`
	Visibility string     `gorm:"type:varchar(20);default:'public'" json:"visibility"`              // YENİ EKLENEN ALAN                                        // Gönderi metni
	ImageURL   string     `json:"image_url"`                                                        // Varsa yüklenen resmin linki
	Likes      int        `gorm:"default:0" json:"likes"`                                           // Toplam beğeni sayısı (Hızlı okuma için)
	Comments   []Comment  `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE;" json:"comments"`   // Postun altındaki yorumlar
	PostLikes  []PostLike `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE;" json:"post_likes"` // Postu kimlerin beğendiği
	CreatedAt  time.Time  `json:"created_at"`
}

// Comment: Gönderilere yapılan yorumlar
type Comment struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"not null;index" json:"post_id"`     // Hangi posta yapıldı?
	UserID    uint      `gorm:"not null" json:"user_id"`           // Kim yaptı?
	User      User      `gorm:"foreignKey:UserID" json:"user"`     // Yorumu yapan kişinin detayları
	Content   string    `gorm:"type:text;not null" json:"content"` // Yorum metni
	CreatedAt time.Time `json:"created_at"`
}

// PostLike: Gönderi beğenilerini takip eden tablo (Çift beğenmeyi önler)
type PostLike struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	PostID uint `gorm:"not null;index" json:"post_id"`
	UserID uint `gorm:"not null;index" json:"user_id"`
}
