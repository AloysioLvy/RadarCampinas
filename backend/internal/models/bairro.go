package models

import "time"

// Bairro representa a localização geográfica de uma denúncia
type Bairro struct {
	BairroID   uint      `json:"id" gorm:"primaryKey;column:id_bairro"`
	Nome       string    `json:"nome" gorm:"size:255;not null"`
	Latitude   string    `json:"latitude" gorm:"not null"`
	Longitude  string    `json:"longitude" gorm:"not null"`
	PesoBairro int       `json:"peso_bairo" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
