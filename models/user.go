package models

import (
	"time"
)

// User: Kullanıcı bilgilerini ve staj tarihlerini tutar
type User struct {
	ID                 uint   `gorm:"primaryKey" json:"id"`
	Phone              string `gorm:"unique;not null" json:"phone"`
	FullNames          string `json:"full_name"`
	Email              string `json:"email"`
	Role               string `gorm:"default:'user'" json:"role"`
	TCNo               string `json:"tc_no"`
	University         string `json:"university"`
	Department         string `json:"department"`
	IsProfileCompleted bool   `gorm:"default:false" json:"is_profile_completed"`

	// STAJ SÜRESİ İÇİN EKLENEN ALANLAR
	InternStartDate *time.Time `json:"intern_start_date"` // Staj başlangıç tarihi
	InternEndDate   *time.Time `json:"intern_end_date"`   // Staj bitiş tarihi
	TotalInternDays int        `json:"total_intern_days"` // Toplam kaç gün staj yapacak (Örn: 40)

	CreatedAt time.Time `json:"created_at"`
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
