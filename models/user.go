package models

import (
	"time"
)

// User: Sisteme giriş yapan herkesin temel kimlik bilgilerini tutar (Auth tablosu)
type User struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	Phone              string    `gorm:"unique;not null" json:"phone"`
	Email              string    `json:"email"`
	FullNames          string    `json:"full_name"`                     // Şirket yetkilisi veya öğrencinin adı
	Role               string    `gorm:"default:'student'" json:"role"` // 'student', 'company', 'admin'
	IsProfileCompleted bool      `gorm:"default:false" json:"is_profile_completed"`
	CreatedAt          time.Time `json:"created_at"`

	// İlişkiler (Has-One) - Bir kullanıcının ya öğrenci profili vardır ya da şirket profili
	StudentProfile *Student `gorm:"foreignKey:UserID" json:"student_profile,omitempty"`
	CompanyProfile *Company `gorm:"foreignKey:UserID" json:"company_profile,omitempty"`
}

// DailyReport: Her gün yazılan staj raporlarını tutar
type DailyReport struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	User      User      `gorm:"foreignKey:UserID"`        // Kullanıcı ile ilişkilendirme
	Content   string    `gorm:"type:text" json:"content"` // Rapor içeriği
	Date      time.Time `gorm:"not null" json:"date"`     // Raporun hangi güne ait olduğu
	CreatedAt time.Time `json:"created_at"`
}

// OTP: Doğrulama kodlarını tutar
type OTP struct {
	ID           uint      `gorm:"primaryKey"`
	Phone        string    `gorm:"index;not null"`
	Code         string    `gorm:"not null"`
	ExpiresAt    time.Time `gorm:"not null"`
	IsUsed       bool      `gorm:"default:false"`
	AttemptCount int       `gorm:"default:0"`
	CreatedAt    time.Time
}
