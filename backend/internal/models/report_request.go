package models

// JSON from front-end
type ReportRequest struct {
	Name        string `json:"name"`
	Latitude    string `json:"latitude"`
	Longitude   string `json:"longitude"`
	CrimeName   string `json:"crime_name"`
	ReportDate  string `json:"report_date"`
	CrimeWeight int    `json:"crime_weight"`
}