package models

import "time"

// Report represents a complete report, relating Neighborhood and Crime
type Report struct {
	ReportID       uint         `json:"report_id" gorm:"primaryKey"`
	NeighborhoodID uint         `json:"neighborhood_id" gorm:"not null"`
	Neighborhood   Neighborhood `json:"neighborhood" gorm:"foreignKey:NeighborhoodID"`
	CrimeID        uint         `json:"crime_id" gorm:"not null"`
	Crime          Crime        `json:"crime" gorm:"foreignKey:CrimeID"`
	ReportDate     string       `json:"report_date" gorm:"not null"`
	CreatedAt      time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}