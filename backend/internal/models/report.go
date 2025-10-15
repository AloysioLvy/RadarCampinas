package models

import "time"

type Report struct {
	ReportID       uint         `json:"report_id" gorm:"primaryKey;column:report_id"`
	NeighborhoodID uint         `json:"neighborhood_id" gorm:"column:neighborhood_id;not null"`
	Neighborhood   Neighborhood `json:"neighborhood" gorm:"foreignKey:NeighborhoodID"`
	CrimeID        uint         `json:"crime_id" gorm:"column:crime_id;not null"`
	Crime          Crime        `json:"crime" gorm:"foreignKey:CrimeID"`
	ReportDate     string       `json:"report_date" gorm:"column:report_date;not null"`
	CreatedAt      time.Time    `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time    `json:"updated_at" gorm:"autoUpdateTime"`
}
