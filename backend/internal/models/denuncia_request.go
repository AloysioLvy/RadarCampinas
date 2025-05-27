package models

// JSON do front-end
type DenunciaRequest struct {
	Nome        string `json:"nome"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
	TipoDeCrime string `json:"tipo_de_crime"`
	DataCrime   string `json:"data_crime"`
	PesoCrime   int    `json:"peso_crime"`
}
