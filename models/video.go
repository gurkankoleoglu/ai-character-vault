package models

import "time"

// Video: Eğitim videolarını ve şirkete özel videoları tutar
type Video struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Title         string    `json:"title"`
	Category      string    `json:"category"`
	Description   string    `gorm:"type:text" json:"description"`
	VideoURL      string    `json:"video_url"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	Duration      string    `json:"duration"`
	Date          time.Time `json:"date"`
	IsCompanyOnly bool      `json:"is_company_only"` // True ise şirkete özel, False ise genel (public)
}
