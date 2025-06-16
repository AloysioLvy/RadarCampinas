package models

import "time"

// Neighborhood represents the geographic location of a report
type Neighborhood struct {
	NeighborhoodID     uint      `json:"neighborhood_id" gorm:"primaryKey"`
	Name               string    `json:"name" gorm:"size:255;not null"`
	Latitude           string    `json:"latitude" gorm:"not null"`
	Longitude          string    `json:"longitude" gorm:"not null"`
	NeighborhoodWeight int       `json:"neighborhood_weight" gorm:"not null"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}