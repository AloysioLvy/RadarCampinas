package models

// JSON do front-end
type DenunciaRequest struct {
	Nome        string `json:"location"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
	TipoDeCrime string `json:"crimeType"`
	DataCrime   string `json:"crimeData"`
	PesoCrime   int    `json:"crimeWeight"`
}
