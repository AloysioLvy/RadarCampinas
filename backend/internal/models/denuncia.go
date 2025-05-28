package models

import "time"

// Denuncia representa uma denúncia completa, relacionando Bairro e Crime
type Denuncia struct {
	DenunciaID uint      `json:"denuncia_id" gorm:"primaryKey"` // descobrir como gerar o id automaticamente
	BairroID   uint      `json:"bairro_id" gorm:"not null"`
	Bairro     Bairro    `json:"bairro" gorm:"foreignKey:BairroID"`
	CrimeID    uint      `json:"crime_id" gorm:"not null"`
	Crime      Crime     `json:"crime" gorm:"foreignKey:CrimeID"`
	DataCrime  string    `json:"data_da_denuncia" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
