package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type Crime struct {
	CrimeID  uint   `json:"crime_id"`
	Name     string `json:"name"`
	Category string `json:"category"` // "Hediondo" ou "Comum"
	Severity int    `json:"severity"`
}

type Neighborhood struct {
	NeighborhoodID     uint      `json:"neighborhood_id"`
	Name               string    `json:"name"`
	Latitude           string    `json:"latitude"`
	Longitude          string    `json:"longitude"`
	NeighborhoodWeight int       `json:"neighborhood_weight"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Report struct {
	ReportID       uint         `json:"report_id"`
	NeighborhoodID uint         `json:"neighborhood_id"`
	Neighborhood   Neighborhood `json:"neighborhood"`
	CrimeID        uint         `json:"crime_id"`
	Crime          Crime        `json:"crime"`
	ReportDate     string       `json:"report_date"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

// Estruturas para dados externos
type WeatherData struct {
	Timestamp time.Time `json:"timestamp"`
	RainMM    float64   `json:"rain_mm"`
	TempC     float64   `json:"temp_c"`
	Humidity  float64   `json:"humidity"`
}

type Holiday struct {
	Date time.Time `json:"date"`
	Name string    `json:"name"`
	Type string    `json:"type"` // nacional, estadual, municipal
}

type Event struct {
	Timestamp  time.Time `json:"timestamp"`
	Name       string    `json:"name"`
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Attendance int       `json:"attendance"`
	Type       string    `json:"type"` // show, esporte, feira, etc
}

// Configuração do gerador
type KnowledgeBaseConfig struct {
	SourceDB       *sql.DB
	TargetDB       *pgxpool.Pool
	CellResolution int // 500 ou 1000 metros
	BatchSize      int
	StartDate      time.Time
	EndDate        time.Time
}

type KnowledgeBaseGenerator struct {
	config *KnowledgeBaseConfig
	logger *log.Logger
}

func NewKnowledgeBaseGenerator(config *KnowledgeBaseConfig) *KnowledgeBaseGenerator {
	return &KnowledgeBaseGenerator{
		config: config,
		logger: log.New(log.Writer(), "[KB-GEN] ", log.LstdFlags),
	}
}

// Map data to schema PostGIS
func (kg *KnowledgeBaseGenerator) mapReportToIncident(report Report) (map[string]interface{}, error) {
	// Convert string cordinates to float64
	lat, err := strconv.ParseFloat(report.Neighborhood.Latitude, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid latitude: %v", err)
	}

	lon, err := strconv.ParseFloat(report.Neighborhood.Longitude, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid longitude: %v", err)
	}

	// bbox
	if !kg.isWithinCampinas(lat, lon) {
		return nil, fmt.Errorf("coordinates outside Campinas bounds")
	}

	// Converter data string para timestamp
	// Convert Data
	reportTime, err := time.Parse("2006-01-02", report.ReportDate)
	if err != nil {
		// Tentar outros formatos comuns
		if reportTime, err = time.Parse("2006-01-02 15:04:05", report.ReportDate); err != nil {
			return nil, fmt.Errorf("invalid date format: %v", err)
		}
	}

	// Mapear categoria do crime
	category := kg.mapCrimeCategory(report.Crime.Name, report.Crime.Category)

	incident := map[string]interface{}{
		"id":           fmt.Sprintf("rpt_%d", report.ReportID),
		"occurred_at":  reportTime,
		"category":     category,
		"severity":     report.Crime.Severity,
		"lat":          lat,
		"lon":          lon,
		"neighborhood": report.Neighborhood.Name,
		"confidence":   kg.calculateConfidence(report),
		"source":       "legacy_reports",
	}

	return incident, nil
}

func (kg *KnowledgeBaseGenerator) isWithinCampinas(lat, lon float64) bool {
	// Bounding box aproximado de Campinas
	// Você pode ajustar estes valores ou usar o polígono oficial
	minLat, maxLat := -23.1, -22.7
	minLon, maxLon := -47.3, -46.8

	return lat >= minLat && lat <= maxLat && lon >= minLon && lon <= maxLon
}

func (kg *KnowledgeBaseGenerator) mapCrimeCategory(crimeName, crimeCategory string) string {
	// Normalizar categoria baseado no nome e categoria original
	crimeName = strings.ToLower(crimeName)
	crimeCategory = strings.ToLower(crimeCategory)

	// Mapear para as categorias do sistema: "Hediondo" ou "Comum"
	hediondos := []string{"homicidio", "latrocinio", "estupro", "sequestro", "trafico"}

	for _, h := range hediondos {
		if strings.Contains(crimeName, h) {
			return "Hediondo"
		}
	}

	if crimeCategory == "hediondo" {
		return "Hediondo"
	}

	return "Comum"
}

func (kg *KnowledgeBaseGenerator) calculateConfidence(report Report) float64 {
	// Score de confiança baseado em fatores como:
	// - Peso do bairro (NeighborhoodWeight)
	// - Idade do dado
	// - Completude dos campos

	confidence := 0.5 // base

	// Fator do peso do bairro (normalizado 0-1)
	if report.Neighborhood.NeighborhoodWeight > 0 {
		confidence += float64(report.Neighborhood.NeighborhoodWeight) / 100.0
	}

	// Penalizar dados muito antigos
	age := time.Since(report.CreatedAt).Hours() / 24 // dias
	if age > 365 {
		confidence *= 0.7
	} else if age > 180 {
		confidence *= 0.85
	}

	// Bonus por completude
	if report.Crime.Severity > 0 {
		confidence += 0.1
	}

	// Limitar entre 0 e 1
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	return confidence
}

// 2. PLANEJAR: Pipeline de transformação e enriquecimento
func (kg *KnowledgeBaseGenerator) GenerateKnowledgeBase(ctx context.Context) error {
	kg.logger.Println("Iniciando geração da base de conhecimento...")

	// Fase 1: Migrar dados históricos
	if err := kg.migrateHistoricalData(ctx); err != nil {
		return fmt.Errorf("erro na migração de dados históricos: %v", err)
	}

	// Fase 2: Gerar grade de células
	if err := kg.generateSpatialGrid(ctx); err != nil {
		return fmt.Errorf("erro na geração da grade espacial: %v", err)
	}

	// Fase 3: Atribuir células aos incidentes
	if err := kg.assignCellsToIncidents(ctx); err != nil {
		return fmt.Errorf("erro na atribuição de células: %v", err)
	}

	// Fase 4: Ingerir dados externos
	if err := kg.ingestExternalData(ctx); err != nil {
		return fmt.Errorf("erro na ingestão de dados externos: %v", err)
	}

	// Fase 5: Gerar features temporais
	if err := kg.generateTemporalFeatures(ctx); err != nil {
		return fmt.Errorf("erro na geração de features: %v", err)
	}

	// Fase 6: Validar qualidade dos dados
	if err := kg.validateDataQuality(ctx); err != nil {
		return fmt.Errorf("erro na validação de qualidade: %v", err)
	}

	kg.logger.Println("Base de conhecimento gerada com sucesso!")
	return nil
}

// 3. EXECUTAR: Implementação das fases
func (kg *KnowledgeBaseGenerator) migrateHistoricalData(ctx context.Context) error {
	kg.logger.Println("Migrando dados históricos...")

	// Query para buscar reports com joins
	query := `
		SELECT r.report_id, r.neighborhood_id, r.crime_id, r.report_date, r.created_at, r.updated_at,
		       n.name as neighborhood_name, n.latitude, n.longitude, n.neighborhood_weight,
		       c.name as crime_name, c.category, c.severity
		FROM reports r
		JOIN neighborhoods n ON r.neighborhood_id = n.neighborhood_id
		JOIN crimes c ON r.crime_id = c.crime_id
		WHERE r.report_date BETWEEN ? AND ?
		ORDER BY r.report_date
	`

	rows, err := kg.config.SourceDB.QueryContext(ctx, query,
		kg.config.StartDate.Format("2006-01-02"),
		kg.config.EndDate.Format("2006-01-02"))
	if err != nil {
		return err
	}
	defer rows.Close()

	batch := make([]map[string]interface{}, 0, kg.config.BatchSize)
	processed := 0

	for rows.Next() {
		var report Report
		var crime Crime
		var neighborhood Neighborhood

		err := rows.Scan(
			&report.ReportID, &report.NeighborhoodID, &report.CrimeID,
			&report.ReportDate, &report.CreatedAt, &report.UpdatedAt,
			&neighborhood.Name, &neighborhood.Latitude, &neighborhood.Longitude,
			&neighborhood.NeighborhoodWeight,
			&crime.Name, &crime.Category, &crime.Severity,
		)
		if err != nil {
			kg.logger.Printf("Erro ao escanear linha: %v", err)
			continue
		}

		report.Neighborhood = neighborhood
		report.Crime = crime

		// Converter para formato do esquema PostGIS
		incident, err := kg.mapReportToIncident(report)
		if err != nil {
			kg.logger.Printf("Erro ao mapear report %d: %v", report.ReportID, err)
			continue
		}

		batch = append(batch, incident)

		// Processar batch quando atingir o tamanho limite
		if len(batch) >= kg.config.BatchSize {
			if err := kg.insertIncidentsBatch(ctx, batch); err != nil {
				return err
			}
			processed += len(batch)
			batch = batch[:0] // limpar slice
			kg.logger.Printf("Processados %d registros...", processed)
		}
	}

	// Processar último batch
	if len(batch) > 0 {
		if err := kg.insertIncidentsBatch(ctx, batch); err != nil {
			return err
		}
		processed += len(batch)
	}

	kg.logger.Printf("Migração concluída: %d incidentes processados", processed)
	return nil
}

func (kg *KnowledgeBaseGenerator) insertIncidentsBatch(ctx context.Context, incidents []map[string]interface{}) error {
	if len(incidents) == 0 {
		return nil
	}

	// Preparar query de inserção em lote
	valueStrings := make([]string, 0, len(incidents))
	valueArgs := make([]interface{}, 0, len(incidents)*9)

	for i, incident := range incidents {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, ST_SetSRID(ST_MakePoint($%d, $%d), 4326)::GEOGRAPHY, $%d, $%d, $%d)",
			i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9))

		valueArgs = append(valueArgs,
			incident["id"],
			incident["occurred_at"],
			incident["category"],
			incident["severity"],
			incident["lon"],
			incident["lat"],
			incident["neighborhood"],
			incident["confidence"],
			incident["source"],
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO curated.incidents (id, occurred_at, category, severity, geom, neighborhood, confidence, source)
		VALUES %s
		ON CONFLICT (id) DO UPDATE SET
			occurred_at = EXCLUDED.occurred_at,
			category = EXCLUDED.category,
			severity = EXCLUDED.severity,
			confidence = EXCLUDED.confidence
	`, strings.Join(valueStrings, ","))

	_, err := kg.config.TargetDB.Exec(ctx, query, valueArgs...)
	return err
}

func (kg *KnowledgeBaseGenerator) generateSpatialGrid(ctx context.Context) error {
	kg.logger.Printf("Gerando grade espacial de %dm...", kg.config.CellResolution)

	// Bounding box de Campinas (ajuste conforme necessário)
	minLon, minLat := -47.3, -23.1
	maxLon, maxLat := -46.8, -22.7

	// Converter resolução de metros para graus (aproximado)
	// 1 grau ≈ 111km, então 500m ≈ 0.0045 graus
	cellSizeDegrees := float64(kg.config.CellResolution) / 111000.0

	cells := make([]map[string]interface{}, 0)
	cellID := 0

	for lon := minLon; lon < maxLon; lon += cellSizeDegrees {
		for lat := minLat; lat < maxLat; lat += cellSizeDegrees {
			cellID++

			// Criar polígono da célula
			cell := map[string]interface{}{
				"cell_id":         fmt.Sprintf("CAMP-%d-%d", kg.config.CellResolution, cellID),
				"cell_resolution": kg.config.CellResolution,
				"city":            "Campinas",
				"min_lon":         lon,
				"min_lat":         lat,
				"max_lon":         lon + cellSizeDegrees,
				"max_lat":         lat + cellSizeDegrees,
			}

			cells = append(cells, cell)
		}
	}

	// Inserir células no banco
	return kg.insertCellsBatch(ctx, cells)
}

func (kg *KnowledgeBaseGenerator) insertCellsBatch(ctx context.Context, cells []map[string]interface{}) error {
	if len(cells) == 0 {
		return nil
	}

	valueStrings := make([]string, 0, len(cells))
	valueArgs := make([]interface{}, 0, len(cells)*7)

	for i, cell := range cells {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, ST_MakeEnvelope($%d, $%d, $%d, $%d, 4326)::GEOGRAPHY)",
			i*7+1, i*7+2, i*7+3, i*7+4, i*7+5, i*7+6, i*7+7))

		valueArgs = append(valueArgs,
			cell["cell_id"],
			cell["cell_resolution"],
			cell["city"],
			cell["min_lon"],
			cell["min_lat"],
			cell["max_lon"],
			cell["max_lat"],
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO curated.cells (cell_id, cell_resolution, city, geom)
		VALUES %s
		ON CONFLICT (cell_id) DO NOTHING
	`, strings.Join(valueStrings, ","))

	_, err := kg.config.TargetDB.Exec(ctx, query, valueArgs...)
	if err != nil {
		return err
	}

	kg.logger.Printf("Inseridas %d células na grade", len(cells))
	return nil
}

func (kg *KnowledgeBaseGenerator) assignCellsToIncidents(ctx context.Context) error {
	kg.logger.Println("Atribuindo células aos incidentes...")

	query := `
		UPDATE curated.incidents 
		SET cell_id = c.cell_id, cell_resolution = c.cell_resolution
		FROM curated.cells c
		WHERE curated.incidents.cell_id IS NULL
		  AND c.cell_resolution = $1
		  AND ST_Contains(c.geom, curated.incidents.geom)
	`

	result, err := kg.config.TargetDB.Exec(ctx, query, kg.config.CellResolution)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	kg.logger.Printf("Atribuídas células a %d incidentes", rowsAffected)

	return nil
}

func (kg *KnowledgeBaseGenerator) ingestExternalData(ctx context.Context) error {
	kg.logger.Println("Ingerindo dados externos...")

	// Simular ingestão de dados de clima
	if err := kg.ingestWeatherData(ctx); err != nil {
		return err
	}

	// Simular ingestão de feriados
	if err := kg.ingestHolidays(ctx); err != nil {
		return err
	}

	// Simular ingestão de eventos
	if err := kg.ingestEvents(ctx); err != nil {
		return err
	}

	return nil
}

func (kg *KnowledgeBaseGenerator) ingestWeatherData(ctx context.Context) error {
	// Simular dados de clima para Campinas
	// Em produção, integra-se com OpenWeatherMap, INMET, etc.

	weatherData := []WeatherData{
		{time.Now().Add(-24 * time.Hour), 5.2, 22.5, 75.0},
		{time.Now().Add(-23 * time.Hour), 0.0, 24.1, 68.0},
		{time.Now().Add(-22 * time.Hour), 12.8, 19.3, 85.0},
		// ... mais dados
	}

	for _, weather := range weatherData {
		query := `
			INSERT INTO external.weather (timestamp, rain_mm, temp_c, humidity, city)
			VALUES ($1, $2, $3, $4, 'Campinas')
			ON CONFLICT (timestamp, city) DO UPDATE SET
				rain_mm = EXCLUDED.rain_mm,
				temp_c = EXCLUDED.temp_c,
				humidity = EXCLUDED.humidity
		`

		_, err := kg.config.TargetDB.Exec(ctx, query,
			weather.Timestamp, weather.RainMM, weather.TempC, weather.Humidity)
		if err != nil {
			return err
		}
	}

	kg.logger.Printf("Ingeridos %d registros de clima", len(weatherData))
	return nil
}

func (kg *KnowledgeBaseGenerator) ingestHolidays(ctx context.Context) error {
	// Calendário de feriados 2025
	holidays := []Holiday{
		{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC), "Ano Novo", "nacional"},
		{time.Date(2025, 4, 21, 0, 0, 0, 0, time.UTC), "Tiradentes", "nacional"},
		{time.Date(2025, 9, 7, 0, 0, 0, 0, time.UTC), "Independência", "nacional"},
		{time.Date(2025, 10, 12, 0, 0, 0, 0, time.UTC), "Nossa Senhora Aparecida", "nacional"},
		{time.Date(2025, 11, 15, 0, 0, 0, 0, time.UTC), "Proclamação da República", "nacional"},
		{time.Date(2025, 12, 25, 0, 0, 0, 0, time.UTC), "Natal", "nacional"},
		// Feriados municipais de Campinas
		{time.Date(2025, 7, 11, 0, 0, 0, 0, time.UTC), "Fundação de Campinas", "municipal"},
	}

	for _, holiday := range holidays {
		query := `
			INSERT INTO external.holidays (date, name, type, city)
			VALUES ($1, $2, $3, 'Campinas')
			ON CONFLICT (date, city) DO NOTHING
		`

		_, err := kg.config.TargetDB.Exec(ctx, query,
			holiday.Date, holiday.Name, holiday.Type)
		if err != nil {
			return err
		}
	}

	kg.logger.Printf("Ingeridos %d feriados", len(holidays))
	return nil
}

func (kg *KnowledgeBaseGenerator) ingestEvents(ctx context.Context) error {
	// Simular eventos em Campinas
	events := []Event{
		{time.Now().Add(-48 * time.Hour), "Show no Estádio", -22.9, -47.1, 15000, "show"},
		{time.Now().Add(-24 * time.Hour), "Feira de Artesanato", -22.91, -47.06, 2000, "feira"},
		{time.Now().Add(-12 * time.Hour), "Jogo de Futebol", -22.89, -47.05, 8000, "esporte"},
	}

	for _, event := range events {
		query := `
			INSERT INTO external.events (timestamp, name, geom, attendance, type, city)
			VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326)::GEOGRAPHY, $5, $6, 'Campinas')
			ON CONFLICT (timestamp, name, city) DO NOTHING
		`

		_, err := kg.config.TargetDB.Exec(ctx, query,
			event.Timestamp, event.Name, event.Longitude, event.Latitude,
			event.Attendance, event.Type)
		if err != nil {
			return err
		}
	}

	kg.logger.Printf("Ingeridos %d eventos", len(events))
	return nil
}

func (kg *KnowledgeBaseGenerator) generateTemporalFeatures(ctx context.Context) error {
	kg.logger.Println("Gerando features temporais...")

	// Gerar features por célula e hora para o período especificado
	current := kg.config.StartDate
	for current.Before(kg.config.EndDate) {
		if err := kg.generateHourlyFeatures(ctx, current); err != nil {
			return err
		}
		current = current.Add(time.Hour)
	}

	return nil
}

func (kg *KnowledgeBaseGenerator) generateHourlyFeatures(ctx context.Context, timestamp time.Time) error {
	// Query complexa para gerar features por célula para uma hora específica
	query := `
		WITH hourly_counts AS (
			SELECT 
				cell_id,
				COUNT(*) as y_count
			FROM curated.incidents
			WHERE occurred_at >= $1 AND occurred_at < $2
			  AND cell_resolution = $3
			GROUP BY cell_id
		),
		lag_features AS (
			SELECT 
				cell_id,
				COALESCE(SUM(CASE WHEN occurred_at >= $4 AND occurred_at < $1 THEN 1 ELSE 0 END), 0) as lag_1h,
				COALESCE(SUM(CASE WHEN occurred_at >= $5 AND occurred_at < $1 THEN 1 ELSE 0 END), 0) as lag_24h,
				COALESCE(SUM(CASE WHEN occurred_at >= $6 AND occurred_at < $1 THEN 1 ELSE 0 END), 0) as lag_7d
			FROM curated.incidents
			WHERE occurred_at >= $6 AND occurred_at < $1
			  AND cell_resolution = $3
			GROUP BY cell_id
		),
		rolling_features AS (
			SELECT 
				cell_id,
				COALESCE(SUM(CASE WHEN occurred_at >= $7 AND occurred_at < $2 THEN 1 ELSE 0 END), 0) as roll_3h_sum,
				COALESCE(SUM(CASE WHEN occurred_at >= $5 AND occurred_at < $2 THEN 1 ELSE 0 END), 0) as roll_24h_sum,
				COALESCE(SUM(CASE WHEN occurred_at >= $6 AND occurred_at < $2 THEN 1 ELSE 0 END), 0) as roll_7d_sum
			FROM curated.incidents
			WHERE occurred_at >= $6 AND occurred_at < $2
			  AND cell_resolution = $3
			GROUP BY cell_id
		),
		weather_features AS (
			SELECT 
				COALESCE(AVG(rain_mm), 0) as weather_rain_mm,
				COALESCE(AVG(temp_c), 20) as weather_temp_c
			FROM external.weather
			WHERE timestamp >= $1 AND timestamp < $2
		),
		holiday_check AS (
			SELECT COUNT(*) > 0 as holiday
			FROM external.holidays
			WHERE date = $1::date
		)
		INSERT INTO features.cell_hourly (
			cell_id, ts, y_count, lag_1h, lag_24h, lag_7d,
			roll_3h_sum, roll_24h_sum, roll_7d_sum,
			dow, hour, weather_rain_mm, weather_temp_c, holiday
		)
		SELECT 
			c.cell_id,
			$1 as ts,
			COALESCE(hc.y_count, 0) as y_count,
			COALESCE(lf.lag_1h, 0) as lag_1h,
			COALESCE(lf.lag_24h, 0) as lag_24h,
			COALESCE(lf.lag_7d, 0) as lag_7d,
			COALESCE(rf.roll_3h_sum, 0) as roll_3h_sum,
			COALESCE(rf.roll_24h_sum, 0) as roll_24h_sum,
			COALESCE(rf.roll_7d_sum, 0) as roll_7d_sum,
			EXTRACT(DOW FROM $1) as dow,
			EXTRACT(HOUR FROM $1) as hour,
			wf.weather_rain_mm,
			wf.weather_temp_c,
			hc_check.holiday
		FROM curated.cells c
		CROSS JOIN weather_features wf
		CROSS JOIN holiday_check hc_check
		LEFT JOIN hourly_counts hc ON c.cell_id = hc.cell_id
		LEFT JOIN lag_features lf ON c.cell_id = lf.cell_id
		LEFT JOIN rolling_features rf ON c.cell_id = rf.cell_id
		WHERE c.cell_resolution = $3
		ON CONFLICT (cell_id, ts) DO UPDATE SET
			y_count = EXCLUDED.y_count,
			lag_1h = EXCLUDED.lag_1h,
			lag_24h = EXCLUDED.lag_24h,
			lag_7d = EXCLUDED.lag_7d,
			roll_3h_sum = EXCLUDED.roll_3h_sum,
			roll_24h_sum = EXCLUDED.roll_24h_sum,
			roll_7d_sum = EXCLUDED.roll_7d_sum,
			weather_rain_mm = EXCLUDED.weather_rain_mm,
			weather_temp_c = EXCLUDED.weather_temp_c,
			holiday = EXCLUDED.holiday
	`

	// Calcular timestamps para lags e janelas
	endHour := timestamp.Add(time.Hour)
	lag1h := timestamp.Add(-time.Hour)
	lag24h := timestamp.Add(-24 * time.Hour)
	lag7d := timestamp.Add(-7 * 24 * time.Hour)
	roll3h := timestamp.Add(-3 * time.Hour)

	_, err := kg.config.TargetDB.Exec(ctx, query,
		timestamp,                // $1
		endHour,                  // $2
		kg.config.CellResolution, // $3
		lag1h,                    // $4
		lag24h,                   // $5
		lag7d,                    // $6
		roll3h,                   // $7
	)

	return err
}

// 4. REVISAR: Validação de qualidade dos dados
func (kg *KnowledgeBaseGenerator) validateDataQuality(ctx context.Context) error {
	kg.logger.Println("Validando qualidade dos dados...")

	// Métricas de qualidade
	metrics := make(map[string]interface{})

	// 1. Cobertura espacial
	var totalCells, cellsWithData int
	err := kg.config.TargetDB.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total_cells,
			COUNT(DISTINCT i.cell_id) as cells_with_data
		FROM curated.cells c
		LEFT JOIN curated.incidents i ON c.cell_id = i.cell_id
		WHERE c.cell_resolution = $1
	`, kg.config.CellResolution).Scan(&totalCells, &cellsWithData)

	if err != nil {
		return err
	}

	spatialCoverage := float64(cellsWithData) / float64(totalCells)
	metrics["spatial_coverage"] = spatialCoverage

	// 2. Cobertura temporal
	var minDate, maxDate time.Time
	var totalHours, hoursWithData int

	err = kg.config.TargetDB.QueryRow(ctx, `
		SELECT 
			MIN(occurred_at) as min_date,
			MAX(occurred_at) as max_date,
			COUNT(DISTINCT DATE_TRUNC('hour', occurred_at)) as hours_with_data
		FROM curated.incidents
	`).Scan(&minDate, &maxDate, &hoursWithData)

	if err != nil {
		return err
	}

	totalHours = int(maxDate.Sub(minDate).Hours())
	temporalCoverage := float64(hoursWithData) / float64(totalHours)
	metrics["temporal_coverage"] = temporalCoverage

	// 3. Taxa de duplicação
	var totalReports, uniqueReports int
	err = kg.config.TargetDB.QueryRow(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(DISTINCT (cell_id, DATE_TRUNC('hour', occurred_at), category)) as unique_reports
		FROM curated.incidents
	`).Scan(&totalReports, &uniqueReports)

	if err != nil {
		return err
	}

	duplicationRate := 1.0 - (float64(uniqueReports) / float64(totalReports))
	metrics["duplication_rate"] = duplicationRate

	// 4. Completude das features
	var featuresWithNulls int
	err = kg.config.TargetDB.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM features.cell_hourly
		WHERE weather_rain_mm IS NULL OR weather_temp_c IS NULL
	`).Scan(&featuresWithNulls)

	if err != nil {
		return err
	}

	var totalFeatures int
	err = kg.config.TargetDB.QueryRow(ctx, `
		SELECT COUNT(*) FROM features.cell_hourly
	`).Scan(&totalFeatures)

	if err != nil {
		return err
	}

	featureCompleteness := 1.0 - (float64(featuresWithNulls) / float64(totalFeatures))
	metrics["feature_completeness"] = featureCompleteness

	// Salvar métricas
	metricsJSON, _ := json.Marshal(metrics)
	_, err = kg.config.TargetDB.Exec(ctx, `
		INSERT INTO analytics.quality_reports (report_date, metrics, created_at)
		VALUES (CURRENT_DATE, $1, NOW())
		ON CONFLICT (report_date) DO UPDATE SET
			metrics = EXCLUDED.metrics,
			created_at = EXCLUDED.created_at
	`, string(metricsJSON))

	if err != nil {
		return err
	}

	// Log das métricas
	kg.logger.Printf("Métricas de qualidade:")
	kg.logger.Printf("  Cobertura espacial: %.2f%%", spatialCoverage*100)
	kg.logger.Printf("  Cobertura temporal: %.2f%%", temporalCoverage*100)
	kg.logger.Printf("  Taxa de duplicação: %.2f%%", duplicationRate*100)
	kg.logger.Printf("  Completude das features: %.2f%%", featureCompleteness*100)

	// Validações críticas
	if spatialCoverage < 0.1 {
		return fmt.Errorf("cobertura espacial muito baixa: %.2f%%", spatialCoverage*100)
	}

	if duplicationRate > 0.5 {
		return fmt.Errorf("taxa de duplicação muito alta: %.2f%%", duplicationRate*100)
	}

	return nil
}
