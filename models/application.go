package models

import (
	"gorm.io/gorm"
)

type Application struct {
	gorm.Model
	FullNames      string `json:"full_name"` // React tarafında full_name bekliyoruz
	Email          string `json:"email"`
	Phone          string `json:"phone"`                           // telefon -> phone (tutarlılık için)
	University     string `json:"university"`                      // okul -> university
	Department     string `json:"department"`                      // bolum -> department
	InternshipType string `json:"intern_type"`                     // stajTuru -> intern_type
	Durum          string `json:"status" gorm:"default:Beklemede"` // durum -> status
	StajTarihleri  string `json:"stajTarihleri"`                   // YENİ EKLENDİ: Başvuru tarihlerini tutacak

	CVUUID     string `json:"cv_uuid"`
	CVPath     string `json:"cv_path"`
	LetterUUID string `json:"letter_uuid"`
	LetterPath string `json:"letter_path"`

	UserID uint `json:"user_id"`
}
