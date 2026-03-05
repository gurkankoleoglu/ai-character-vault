package models

import (
	"time"
)

type Announcement struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string    `gorm:"type:varchar(255);not null" json:"title"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	Subject   string    `gorm:"type:varchar(255)" json:"subject"`
	Date      time.Time `gorm:"not null" json:"date"`
	Type      string    `gorm:"type:varchar(50);not null" json:"type"`
	ImageURL  string    `gorm:"type:varchar(500)" json:"image_url"` // YENİ EKLENDİ: Resim yolunu tutacak alan
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
