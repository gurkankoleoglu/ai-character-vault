package models

import "time"

// JobPosting: Şirketlerin açtığı staj/iş ilanlarını tutar
type JobPosting struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	CompanyID        uint      `gorm:"not null" json:"company_id"` // Hangi şirket açtı?
	Company          Company   `gorm:"foreignKey:CompanyID" json:"company"`
	Title            string    `json:"title"`                              // İlan Başlığı (Örn: Go Backend Stajyeri)
	Description      string    `gorm:"type:text" json:"description"`       // İlan Detayları
	Requirements     string    `gorm:"type:text" json:"requirements"`      // Aranan Nitelikler
	Benefits         string    `gorm:"type:text" json:"benefits"`          // YENİ: Sağlanan İmkanlar
	WorkType         string    `json:"work_type"`                          // YENİ: Çalışma Tipi (full-time, remote vb.)
	InternshipType   string    `json:"internship_type"`                    // İlan Kategorisi (job, internship)
	Capacity         int       `json:"capacity"`                           // Kaç kişi alınacak (Kontenjan)
	Location         string    `json:"location"`                           // Ankara, Remote vb.
	IsActive         bool      `gorm:"default:true" json:"is_active"`      // İlan yayında mı?
	ApplicationCount int       `gorm:"default:0" json:"application_count"` // Kaç kişi başvurdu
	CreatedAt        time.Time `json:"created_at"`
}
