package models

type Crime struct {
	IDCrime 	int64   `gorm:"column:id_crime;primaryKey" json:"id_crime"`
	NomeCrime	*string `gorm:"column:nome_crime" json:nome_crime"`
	PesoCrime	*int16 	`gorm:"column:peso_crime" json:peso_crime"`
}

func (Crime) TableName() string {
	return "Crimes"
}