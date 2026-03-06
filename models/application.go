package models

import (
	"gorm.io/gorm"
)

type Application struct {
	gorm.Model
	// YENİ EKLENEN: Başvurunun hangi ilana yapıldığını tutar
	JobPostingID uint       `json:"job_posting_id"`
	JobPosting   JobPosting `gorm:"foreignKey:JobPostingID" json:"job_posting"`

	FullNames      string `json:"full_name"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	University     string `json:"university"`
	Department     string `json:"department"`
	InternshipType string `json:"intern_type"`
	StajTarihleri  string `json:"stajTarihleri"`
	Durum          string `json:"status" gorm:"default:Beklemede"`

	CVUUID     string `json:"cv_uuid"`
	CVPath     string `json:"cv_path"`
	LetterUUID string `json:"letter_uuid"`
	LetterPath string `json:"letter_path"`

	UserID uint `json:"user_id"`
}
