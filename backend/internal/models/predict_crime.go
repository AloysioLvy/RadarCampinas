package models

type PredictCrime struct {
	PredictCrimeID     uint      `json:"predict_crime_id" gorm:"primaryKey;column:predict_crime_id"`
	Neighborhood     string      `json:"neighborhood" gorm:"column:neighborhood;size:255;not null"`
	Risk_Level int       `json:"risk_level" gorm:"column:risk_level;not null"`
}