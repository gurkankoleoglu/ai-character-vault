package models

import "time"

// Student: Sadece 'student' rolündeki kullanıcıların ekstra verilerini tutar
type Student struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	UserID         uint   `gorm:"unique;not null" json:"user_id"`
	TCNo           string `json:"tc_no"`
	University     string `json:"university"`
	Department     string `json:"department"`
	StudentCertUrl string `json:"student_cert_url"`
	IsCertApproved bool   `gorm:"default:false" json:"is_cert_approved"`

	// YENİ EKLENEN PROFİL ALANLARI
	Skills string `json:"skills"` // Örn: "React, Python, SQL"
	Github string `json:"github"` // Örn: "https://github.com/kullanici"

	InternStartDate *time.Time `json:"intern_start_date"`
	InternEndDate   *time.Time `json:"intern_end_date"`
	TotalInternDays int        `json:"total_intern_days"`
}
