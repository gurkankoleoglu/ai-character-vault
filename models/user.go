package models

import (
	"time"
)

// User: Sisteme giriş yapan herkesin temel kimlik bilgilerini tutar
type User struct {
	ID                 uint   `gorm:"primaryKey" json:"id"`
	Phone              string `gorm:"unique;not null" json:"phone"`
	Email              string `json:"email"`
	FullNames          string `json:"full_name"`
	Role               string `gorm:"default:'student'" json:"role"`
	IsProfileCompleted bool   `gorm:"default:false" json:"is_profile_completed"`
	IsOnline           bool   `gorm:"default:false" json:"is_online"` // YENİ ALAN
	// YENİ EKLENEN PROFİL ALANLARI
	Password  string `json:"-"`
	Location  string `json:"location"`
	Bio       string `gorm:"type:text" json:"bio"`
	AvatarURL string `json:"avatar_url"`

	CreatedAt time.Time `json:"created_at"`

	StudentProfile *Student `gorm:"foreignKey:UserID" json:"student_profile,omitempty"`
	CompanyProfile *Company `gorm:"foreignKey:UserID" json:"company_profile,omitempty"`
}

// DailyReport: Her gün yazılan staj raporlarını tutar
type DailyReport struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `json:"user_id"`
	User      User      `gorm:"foreignKey:UserID"`
	Content   string    `gorm:"type:text" json:"content"`
	Date      time.Time `gorm:"not null" json:"date"`
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
