package models

import "time"

// Crime representa o tipo de crime e seu peso
type Crime struct {
	CrimeID     uint      `json:"id" gorm:"primaryKey"`
	TipoDeCrime string    `json:"tipo_de_crime" gorm:"size:255;not null"`
	PesoCrime   int       `json:"peso_crime" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
