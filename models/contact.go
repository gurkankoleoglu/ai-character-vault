package models

import "gorm.io/gorm"

type ContactMessage struct {
	gorm.Model
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}
