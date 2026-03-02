package models

import (
	"time"
)

type User struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Phone     string `gorm:"unique;not null" json:"phone"`
	FullNames string `json:"full_name"`
	Email     string `json:"email"`
	// Handler'da hata veren eksik alanları buraya ekledik:
	Role               string    `gorm:"default:'user'" json:"role"`
	TCNo               string    `json:"tc_no"`
	University         string    `json:"university"`
	Department         string    `json:"department"`
	IsProfileCompleted bool      `gorm:"default:false" json:"is_profile_completed"`
	CreatedAt          time.Time `json:"created_at"`
}

type OTP struct {
	ID           uint      `gorm:"primaryKey"`
	Phone        string    `gorm:"index;not null"`
	Code         string    `gorm:"not null"`
	ExpiresAt    time.Time `gorm:"not null"`
	IsUsed       bool      `gorm:"default:false"`
	AttemptCount int       `gorm:"default:0"`
	CreatedAt    time.Time
}
