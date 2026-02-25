package models

import (
	"gorm.io/gorm"
)

type Application struct {
	gorm.Model
	FullNames      string `json:"adSoyad"`
	Email          string `json:"email"`
	Phone          string `json:"telefon"`
	University     string `json:"okul"`
	Department     string `json:"bolum"`
	InternshipType string `json:"stajTuru"`

	// UUID ve Dosya Yolları
	CVUUID     string `json:"cv_uuid"` // CV için benzersiz ID
	CVPath     string `json:"cv_path"`
	LetterUUID string `json:"letter_uuid"` // Motivasyon mektubu için benzersiz ID
	LetterPath string `json:"letter_path"`

	// Kullanıcı İlişkisi
	UserID uint `json:"user_id"` // Hangi kullanıcı yüklediyse onun ID'si
}
