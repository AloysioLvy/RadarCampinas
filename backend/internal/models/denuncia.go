package models

import "time"

type Denuncia struct {
	IDDenuncia	int64		`gorm:"column:id_denuncia;primaryKey" json:"id_denuncia"`
	IDCrime		int64		`gorm:"column:id_crime" json:"id_crime"`
	IDBairro	int64		`gorm:"column:id_bairro" json:"id_bairro"`
	Data 		time.Time 	`gorm:"column:data" json:"data"`
}

func (Denuncia) TableName() string {
	return "Denuncia"
}