package models

// Company: Sadece 'company' rolündeki kullanıcıların şirket verilerini tutar
type Company struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	UserID      uint   `gorm:"unique;not null" json:"user_id"` // User tablosuyla 1-1 ilişki
	CompanyName string `json:"company_name"`
	TaxNumber   string `json:"tax_number"` // Vergi No (Sahte şirketleri önlemek için)
	Industry    string `json:"industry"`   // Sektör (Örn: Yazılım, Tasarım)
	LogoUrl     string `json:"logo_url"`
	Website     string `json:"website"`
	Address     string `json:"address"`
	IsApproved  bool   `gorm:"default:false" json:"is_approved"` // Admin şirketi doğruladı mı?
}
