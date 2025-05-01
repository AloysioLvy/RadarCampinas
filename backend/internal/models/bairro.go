package models

type Bairro struct {
	IDBairro 	int64 	`gorm:"column:id_bairro;primaryKey" json:"id_bairro"`
	Latitude	string	`gorm:"column:latitude" json:"latitude"`
	Longitude	string	`gorm:"column:longitude" json:"longitude"`
	Nome 		string	`gorm:"column:nome" json:"nome"`
	PesoBairro	int64	`gorm:"column:peso_bairro" json:"peso_bairro"`
}

func (Bairro) TableName() string {
	return "Bairros"
}