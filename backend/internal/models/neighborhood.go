package models

import "time"

type Neighborhood struct {
	NeighborhoodID     uint      `json:"neighborhood_id" gorm:"primaryKey;column:neighborhood_id"`
	Name               string    `json:"name" gorm:"column:name;size:255;not null"`
	Latitude           string    `json:"latitude" gorm:"column:latitude;not null"`
	Longitude          string    `json:"longitude" gorm:"column:longitude;not null"`
	NeighborhoodWeight int       `json:"neighborhood_weight" gorm:"column:neighborhood_weight;not null"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
