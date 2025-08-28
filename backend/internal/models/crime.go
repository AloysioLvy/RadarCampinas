package models

import "time"

// Crime represents the type of crime and its weight
type Crime struct {
	CrimeID    uint      `json:"crime_id" gorm:"primaryKey"`
	CrimeName  string    `json:"crime_name" gorm:"size:255;not null"`
	CrimeWeight int      `json:"crime_weight" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
