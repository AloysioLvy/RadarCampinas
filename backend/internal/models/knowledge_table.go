package models

import "time"

// ============================================================================
// CURATED SCHEMA
// ============================================================================

// CuratedIncident representa um incidente criminal curado
type CuratedIncident struct {
	ID             string    `json:"id" gorm:"primaryKey;column:id;size:50"`
	OccurredAt     time.Time `json:"occurred_at" gorm:"column:occurred_at;not null;index"`
	Category       string    `json:"category" gorm:"column:category;size:50;not null;index"`
	Severity       int       `json:"severity" gorm:"column:severity;not null;check:severity >= 1 AND severity <= 10;index"`
	Latitude       float64   `json:"latitude" gorm:"column:latitude;type:decimal(10,8);not null;index:idx_lat_lng"`
	Longitude      float64   `json:"longitude" gorm:"column:longitude;type:decimal(11,8);not null;index:idx_lat_lng"`
	Neighborhood   string    `json:"neighborhood" gorm:"column:neighborhood;size:100"`
	Confidence     float64   `json:"confidence" gorm:"column:confidence;type:float;check:confidence >= 0 AND confidence <= 1"`
	Source         string    `json:"source" gorm:"column:source;size:50;default:'legacy_reports'"`
	CellID         *string   `json:"cell_id" gorm:"column:cell_id;size:50;index"`
	CellResolution *int      `json:"cell_resolution" gorm:"column:cell_resolution"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CuratedIncident) TableName() string {
	return "curated_incidents"
}

// CuratedCell representa uma célula da grade espacial
// CuratedCell representa uma célula da grade espacial
type CuratedCell struct {
	CellID         string    `json:"cell_id" gorm:"primaryKey;column:cell_id;size:50"`
	CellResolution int       `json:"cell_resolution" gorm:"column:cell_resolution;not null;index"`
	City           string    `json:"city" gorm:"column:city;size:50;default:'Campinas';index"`
	CenterLat      float64   `json:"center_lat" gorm:"column:center_lat;type:decimal(10,8);not null;index:idx_center"`
	CenterLng      float64   `json:"center_lng" gorm:"column:center_lng;type:decimal(11,8);not null;index:idx_center"`

	// Guardar JSON como string em nvarchar(max) no SQL Server
	BoundsJSON *string `json:"bounds_json" gorm:"column:bounds_json;type:nvarchar(max)"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (CuratedCell) TableName() string {
	return "curated_cells"
}

// ============================================================================
// EXTERNAL SCHEMA
// ============================================================================

// ExternalHoliday representa um feriado
type ExternalHoliday struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	Date      time.Time `json:"date" gorm:"column:date;type:date;not null;uniqueIndex:unique_date_city;index"`
	Name      string    `json:"name" gorm:"column:name;size:100;not null"`
	Type      string    `json:"type" gorm:"column:type;size:50"`
	City      string    `json:"city" gorm:"column:city;size:50;default:'Campinas';uniqueIndex:unique_date_city;index"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (ExternalHoliday) TableName() string {
	return "external_holidays"
}

// ============================================================================
// FEATURES SCHEMA
// ============================================================================

// FeaturesCellHourly representa features por célula e hora
type FeaturesCellHourly struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	CellID    string    `json:"cell_id" gorm:"column:cell_id;size:50;not null;uniqueIndex:unique_cell_ts;index"`
	Ts        time.Time `json:"ts" gorm:"column:ts;not null;uniqueIndex:unique_cell_ts;index"`
	
	// Target variable
	YCount int `json:"y_count" gorm:"column:y_count;default:0"`
	
	// Lag features
	Lag1h  int `json:"lag_1h" gorm:"column:lag_1h;default:0"`
	Lag24h int `json:"lag_24h" gorm:"column:lag_24h;default:0"`
	Lag7d  int `json:"lag_7d" gorm:"column:lag_7d;default:0"`
	
	// Rolling window features
	Roll3hSum  int      `json:"roll_3h_sum" gorm:"column:roll_3h_sum;default:0"`
	Roll24hSum int      `json:"roll_24h_sum" gorm:"column:roll_24h_sum;default:0"`
	Roll7dSum  int      `json:"roll_7d_sum" gorm:"column:roll_7d_sum;default:0"`
	Roll7dAvg  *float64 `json:"roll_7d_avg" gorm:"column:roll_7d_avg;type:float"`
	Roll7dStd  *float64 `json:"roll_7d_std" gorm:"column:roll_7d_std;type:float"`
	
	// Temporal features
	Dow             *int  `json:"dow" gorm:"column:dow;index"`
	Hour            *int  `json:"hour" gorm:"column:hour;index"`
	IsWeekend       *bool `json:"is_weekend" gorm:"column:is_weekend"`
	IsBusinessHours *bool `json:"is_business_hours" gorm:"column:is_business_hours"`
	
	// Calendar features
	Holiday          bool `json:"holiday" gorm:"column:holiday;default:false"`
	DayBeforeHoliday bool `json:"day_before_holiday" gorm:"column:day_before_holiday;default:false"`
	DayAfterHoliday  bool `json:"day_after_holiday" gorm:"column:day_after_holiday;default:false"`
	
	// Spatial features
	NeighborAvgCrime *float64 `json:"neighbor_avg_crime" gorm:"column:neighbor_avg_crime;type:float"`
	
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (FeaturesCellHourly) TableName() string {
	return "features_cell_hourly"
}

// ============================================================================
// ANALYTICS SCHEMA
// ============================================================================

// AnalyticsQualityReport representa um relatório de qualidade
// AnalyticsQualityReport representa um relatório de qualidade
type AnalyticsQualityReport struct {
	ID         uint      `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	ReportDate time.Time `json:"report_date" gorm:"column:report_date;type:date;uniqueIndex;not null"`

	// Trocar type:json por nvarchar(max)
	Metrics string `json:"metrics" gorm:"column:metrics;type:nvarchar(max);not null"`

	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
func (AnalyticsQualityReport) TableName() string {
	return "analytics_quality_reports"
}

// AnalyticsPipelineLog representa um log de execução do pipeline
// AnalyticsPipelineLog representa um log de execução do pipeline
type AnalyticsPipelineLog struct {
	ID                   uint       `json:"id" gorm:"primaryKey;autoIncrement;column:id"`
	ExecutionID          string     `json:"execution_id" gorm:"column:execution_id;size:36;uniqueIndex;not null"`
	StartedAt            time.Time  `json:"started_at" gorm:"column:started_at;not null;index"`
	FinishedAt           *time.Time `json:"finished_at" gorm:"column:finished_at"`
	Status               string     `json:"status" gorm:"column:status;size:20;index"`
	Phase                string     `json:"phase" gorm:"column:phase;size:50"`
	RecordsProcessed     *int       `json:"records_processed" gorm:"column:records_processed"`

	// text -> nvarchar(max)
	ErrorMessage *string `json:"error_message" gorm:"column:error_message;type:nvarchar(max)"`

	ExecutionTimeSeconds *int      `json:"execution_time_seconds" gorm:"column:execution_time_seconds"`
	CreatedAt            time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (AnalyticsPipelineLog) TableName() string {
	return "analytics_pipeline_logs"
}

// ============================================================================
// MIGRATIONS TABLE
// ============================================================================

// SchemaMigration representa uma versão de migration aplicada
type SchemaMigration struct {
	Version   string    `json:"version" gorm:"primaryKey;column:version;size:50"`
	AppliedAt time.Time `json:"applied_at" gorm:"autoCreateTime"`
}

func (SchemaMigration) TableName() string {
	return "schema_migrations"
}