package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Crime struct {
	CrimeID     uint   `json:"crime_id"`
	CrimeName   string `json:"crime_name"`
	CrimeWeight int    `json:"crime_weight"`
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

type KnowledgeBaseConfig struct {
	SourceDB       *sql.DB
	TargetDB       *sql.DB // Mesmo banco agora
	CellResolution int
	BatchSize      int
	StartDate      time.Time
	EndDate        time.Time
}

type KnowledgeBaseGenerator struct {
	config      *KnowledgeBaseConfig
	logger      *log.Logger
	executionID string
}

func NewKnowledgeBaseGenerator(config *KnowledgeBaseConfig) *KnowledgeBaseGenerator {
	return &KnowledgeBaseGenerator{
		config:      config,
		logger:      log.New(os.Stdout, "[KB-GEN] ", log.LstdFlags|log.Lmsgprefix),
		executionID: fmt.Sprintf("exec_%d", time.Now().Unix()),
	}
}

// ============================================================================
// PIPELINE PRINCIPAL
// ============================================================================

func (kg *KnowledgeBaseGenerator) GenerateKnowledgeBase(ctx context.Context) error {
	kg.logger.Println("🚀 Iniciando geração da base de conhecimento...")
	kg.logger.Printf("📋 Execution ID: %s", kg.executionID)
	kg.logger.Printf("📅 Período: %s até %s",
		kg.config.StartDate.Format("2006-01-02"),
		kg.config.EndDate.Format("2006-01-02"))

	startTime := time.Now()
	db := kg.config.SourceDB // Usar apenas um banco

	// Fase 1: Migrar dados históricos
	kg.logger.Println("📊 Fase 1: Migrando dados históricos...")
	if err := kg.migrateHistoricalData(ctx, db); err != nil {
		return fmt.Errorf("❌ erro na migração: %v", err)
	}

	// Fase 2: Gerar grade espacial
	kg.logger.Println("🗺️  Fase 2: Gerando grade espacial...")
	if err := kg.generateSpatialGrid(ctx, db); err != nil {
		return fmt.Errorf("❌ erro na grade espacial: %v", err)
	}

	// Fase 3: Atribuir células aos incidentes
	kg.logger.Println("🎯 Fase 3: Atribuindo células aos incidentes...")
	if err := kg.assignCellsToIncidents(ctx, db); err != nil {
		return fmt.Errorf("❌ erro na atribuição de células: %v", err)
	}

	// Fase 4: Gerar features temporais
	kg.logger.Println("⚙️  Fase 4: Gerando features temporais...")
	if err := kg.generateTemporalFeatures(ctx, db); err != nil {
		return fmt.Errorf("❌ erro na geração de features: %v", err)
	}

	// Fase 5: Validar qualidade
	kg.logger.Println("✓ Fase 5: Validando qualidade dos dados...")
	if err := kg.validateDataQuality(ctx, db); err != nil {
		return fmt.Errorf("❌ erro na validação: %v", err)
	}

	executionTime := time.Since(startTime)
	kg.logger.Printf("✅ Base de conhecimento gerada com sucesso em %s!", executionTime)

	return nil
}

// ============================================================================
// FASE 1: MIGRAÇÃO DE DADOS HISTÓRICOS
// ============================================================================

func (kg *KnowledgeBaseGenerator) migrateHistoricalData(ctx context.Context, db *sql.DB) error {
	query := `
		SELECT r.report_id, r.neighborhood_id, r.crime_id, r.report_date, r.created_at, r.updated_at,
			n.name as neighborhood_name, n.latitude, n.longitude, n.neighborhood_weight,
			c.crime_name, c.crime_weight
		FROM reports r
		JOIN neighborhoods n ON r.neighborhood_id = n.neighborhood_id
		JOIN crimes c ON r.crime_id = c.crime_id
		WHERE r.report_date BETWEEN ? AND ?
		ORDER BY r.report_date
	`

	rows, err := db.QueryContext(ctx, query,
		kg.config.StartDate.Format("2006-01-02"),
		kg.config.EndDate.Format("2006-01-02"))
	if err != nil {
		return fmt.Errorf("erro na query: %v", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			kg.logger.Printf("⚠️  Erro ao fechar rows: %v", err)
		}
	}()

	processed := 0
	skipped := 0

	for rows.Next() {
		var report Report
		var crime Crime
		var neighborhood Neighborhood

		err := rows.Scan(
			&report.ReportID, &report.NeighborhoodID, &report.CrimeID,
			&report.ReportDate, &report.CreatedAt, &report.UpdatedAt,
			&neighborhood.Name, &neighborhood.Latitude, &neighborhood.Longitude,
			&neighborhood.NeighborhoodWeight,
			&crime.CrimeName, &crime.CrimeWeight,
		)
		if err != nil {
			kg.logger.Printf("⚠️  Erro ao escanear linha: %v", err)
			skipped++
			continue
		}

		report.Neighborhood = neighborhood
		report.Crime = crime

		if err := kg.insertIncident(ctx, db, report); err != nil {
			kg.logger.Printf("⚠️  Erro ao inserir report %d: %v", report.ReportID, err)
			skipped++
			continue
		}

		processed++
		if processed%100 == 0 {
			kg.logger.Printf("  ➜ Processados %d registros...", processed)
		}
	}

	kg.logger.Printf("✅ Migração concluída: %d incidentes processados, %d ignorados", processed, skipped)
	return nil
}

func (kg *KnowledgeBaseGenerator) insertIncident(ctx context.Context, db *sql.DB, report Report) error {
	lat, err := strconv.ParseFloat(report.Neighborhood.Latitude, 64)
	if err != nil {
		return fmt.Errorf("latitude inválida: %v", err)
	}

	lon, err := strconv.ParseFloat(report.Neighborhood.Longitude, 64)
	if err != nil {
		return fmt.Errorf("longitude inválida: %v", err)
	}

	// Validar coordenadas de Campinas
	if lat < -23.1 || lat > -22.7 || lon < -47.3 || lon > -46.8 {
		return fmt.Errorf("coordenadas fora de Campinas")
	}

	reportTime, err := time.Parse("2006-01-02", report.ReportDate)
	if err != nil {
		if reportTime, err = time.Parse("2006-01-02 15:04:05", report.ReportDate); err != nil {
			return fmt.Errorf("formato de data inválido: %v", err)
		}
	}

	category := kg.mapCrimeCategory(report.Crime.CrimeName)
	severity := report.Crime.CrimeWeight
	confidence := kg.calculateConfidence(report)

	query := `
		INSERT INTO curated_incidents 
		(id, occurred_at, category, severity, latitude, longitude, neighborhood, confidence, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			occurred_at = VALUES(occurred_at),
			category = VALUES(category),
			severity = VALUES(severity),
			confidence = VALUES(confidence)
	`

	_, err = db.ExecContext(ctx, query,
		fmt.Sprintf("rpt_%d", report.ReportID),
		reportTime,
		category,
		severity,
		lat,
		lon,
		report.Neighborhood.Name,
		confidence,
		"legacy_reports",
	)

	return err
}

func (kg *KnowledgeBaseGenerator) mapCrimeCategory(crimeName string) string {
	crimeName = strings.ToLower(crimeName)
	hediondos := []string{"homicidio", "homicídio", "latrocinio", "latrocínio", "estupro", "sequestro", "trafico", "tráfico"}

	for _, h := range hediondos {
		if strings.Contains(crimeName, h) {
			return "Hediondo"
		}
	}
	return "Comum"
}

func (kg *KnowledgeBaseGenerator) calculateConfidence(report Report) float64 {
	confidence := 0.5

	if report.Neighborhood.NeighborhoodWeight > 0 {
		confidence += float64(report.Neighborhood.NeighborhoodWeight) / 100.0
	}

	age := time.Since(report.CreatedAt).Hours() / 24
	if age > 365 {
		confidence *= 0.7
	} else if age > 180 {
		confidence *= 0.85
	}

	if report.Crime.CrimeWeight > 0 {
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	return confidence
}

// ============================================================================
// FASE 2: GRADE ESPACIAL
// ============================================================================

func (kg *KnowledgeBaseGenerator) generateSpatialGrid(ctx context.Context, db *sql.DB) error {
	minLon, minLat := -47.3, -23.1
	maxLon, maxLat := -46.8, -22.7

	cellSizeDegrees := float64(kg.config.CellResolution) / 111000.0

	cellID := 0
	processed := 0

	for lon := minLon; lon < maxLon; lon += cellSizeDegrees {
		for lat := minLat; lat < maxLat; lat += cellSizeDegrees {
			cellID++

			centerLat := lat + cellSizeDegrees/2
			centerLon := lon + cellSizeDegrees/2

			query := `
				INSERT INTO curated_cells 
				(cell_id, cell_resolution, city, center_lat, center_lng)
				VALUES (?, ?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE cell_id = cell_id
			`

			_, err := db.ExecContext(ctx, query,
				fmt.Sprintf("CAMP-%d-%d", kg.config.CellResolution, cellID),
				kg.config.CellResolution,
				"Campinas",
				centerLat,
				centerLon,
			)

			if err != nil {
				return err
			}

			processed++
		}
	}

	kg.logger.Printf("✅ Grade espacial gerada: %d células", processed)
	return nil
}

// ============================================================================
// FASE 3: ATRIBUIR CÉLULAS
// ============================================================================

func (kg *KnowledgeBaseGenerator) assignCellsToIncidents(ctx context.Context, db *sql.DB) error {
	// Buscar todas as células
	cellsQuery := `SELECT cell_id, center_lat, center_lng FROM curated_cells WHERE cell_resolution = ?`
	cellRows, err := db.QueryContext(ctx, cellsQuery, kg.config.CellResolution)
	if err != nil {
		return err
	}
	defer func() {
		if err := cellRows.Close(); err != nil {
			kg.logger.Printf("⚠️  Erro ao fechar cellRows: %v", err)
		}
	}()

	type Cell struct {
		ID        string
		CenterLat float64
		CenterLng float64
	}

	cells := []Cell{}
	for cellRows.Next() {
		var c Cell
		if err := cellRows.Scan(&c.ID, &c.CenterLat, &c.CenterLng); err != nil {
			continue
		}
		cells = append(cells, c)
	}

	// Buscar incidentes sem célula
	incidentsQuery := `SELECT id, latitude, longitude FROM curated_incidents WHERE cell_id IS NULL`
	incidentRows, err := db.QueryContext(ctx, incidentsQuery)
	if err != nil {
		return err
	}
	defer func() {
		if err := incidentRows.Close(); err != nil {
			kg.logger.Printf("⚠️  Erro ao fechar incidentRows: %v", err)
		}
	}()

	updated := 0
	cellSizeDegrees := float64(kg.config.CellResolution) / 111000.0

	for incidentRows.Next() {
		var incidentID string
		var lat, lon float64

		if err := incidentRows.Scan(&incidentID, &lat, &lon); err != nil {
			continue
		}

		// Encontrar célula mais próxima
		for _, cell := range cells {
			if lat >= cell.CenterLat-cellSizeDegrees/2 && lat < cell.CenterLat+cellSizeDegrees/2 &&
				lon >= cell.CenterLng-cellSizeDegrees/2 && lon < cell.CenterLng+cellSizeDegrees/2 {

				updateQuery := `UPDATE curated_incidents SET cell_id = ?, cell_resolution = ? WHERE id = ?`
				_, err := db.ExecContext(ctx, updateQuery, cell.ID, kg.config.CellResolution, incidentID)
				if err == nil {
					updated++
				}
				break
			}
		}
	}

	kg.logger.Printf("✅ Atribuídas células a %d incidentes", updated)
	return nil
}

// ============================================================================
// FASE 4: FEATURES TEMPORAIS
// ============================================================================

func (kg *KnowledgeBaseGenerator) generateTemporalFeatures(ctx context.Context, db *sql.DB) error {
	current := kg.config.StartDate
	hoursProcessed := 0

	for current.Before(kg.config.EndDate) {
		if err := kg.generateHourlyFeatures(ctx, db, current); err != nil {
			kg.logger.Printf("⚠️  Erro ao processar hora %s: %v", current, err)
		}
		hoursProcessed++
		current = current.Add(time.Hour)

		if hoursProcessed%100 == 0 {
			kg.logger.Printf("  ➜ Processadas %d horas...", hoursProcessed)
		}
	}

	kg.logger.Printf("✅ Features temporais geradas: %d horas processadas", hoursProcessed)
	return nil
}

func (kg *KnowledgeBaseGenerator) generateHourlyFeatures(ctx context.Context, db *sql.DB, timestamp time.Time) error {
	// Buscar todas as células
	cellsQuery := `SELECT cell_id FROM curated_cells WHERE cell_resolution = ?`
	rows, err := db.QueryContext(ctx, cellsQuery, kg.config.CellResolution)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			kg.logger.Printf("⚠️  Erro ao fechar rows: %v", err)
		}
	}()

	endHour := timestamp.Add(time.Hour)

	for rows.Next() {
		var cellID string
		if err := rows.Scan(&cellID); err != nil {
			continue
		}

		// Contar crimes nessa hora
		var yCount int
		countQuery := `
			SELECT COUNT(*) FROM curated_incidents 
			WHERE cell_id = ? AND occurred_at >= ? AND occurred_at < ?
		`
		if err := db.QueryRowContext(ctx, countQuery, cellID, timestamp, endHour).Scan(&yCount); err != nil {
			kg.logger.Printf("⚠️  Erro ao contar crimes: %v", err)
			continue
		}

		// Calcular lags
		lag1h := timestamp.Add(-time.Hour)
		lag24h := timestamp.Add(-24 * time.Hour)
		lag7d := timestamp.Add(-7 * 24 * time.Hour)

		var lag1hCount, lag24hCount, lag7dCount int
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM curated_incidents WHERE cell_id = ? AND occurred_at >= ? AND occurred_at < ?`,
			cellID, lag1h, timestamp).Scan(&lag1hCount); err != nil {
			kg.logger.Printf("⚠️  Erro ao calcular lag 1h: %v", err)
		}
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM curated_incidents WHERE cell_id = ? AND occurred_at >= ? AND occurred_at < ?`,
			cellID, lag24h, timestamp).Scan(&lag24hCount); err != nil {
			kg.logger.Printf("⚠️  Erro ao calcular lag 24h: %v", err)
		}
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM curated_incidents WHERE cell_id = ? AND occurred_at >= ? AND occurred_at < ?`,
			cellID, lag7d, timestamp).Scan(&lag7dCount); err != nil {
			kg.logger.Printf("⚠️  Erro ao calcular lag 7d: %v", err)
		}

		// Verificar feriado
		var isHoliday bool
		if err := db.QueryRowContext(ctx, `SELECT COUNT(*) > 0 FROM external_holidays WHERE date = ?`, timestamp.Format("2006-01-02")).Scan(&isHoliday); err != nil {
			kg.logger.Printf("⚠️  Erro ao verificar feriado: %v", err)
		}

		dow := int(timestamp.Weekday())
		hour := timestamp.Hour()
		isWeekend := dow == 0 || dow == 6
		isBusinessHours := hour >= 8 && hour <= 18

		// Inserir feature
		insertQuery := `
			INSERT INTO features_cell_hourly 
			(cell_id, ts, y_count, lag_1h, lag_24h, lag_7d, dow, hour, holiday, is_weekend, is_business_hours)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE
				y_count = VALUES(y_count),
				lag_1h = VALUES(lag_1h),
				lag_24h = VALUES(lag_24h),
				lag_7d = VALUES(lag_7d)
		`

		_, err := db.ExecContext(ctx, insertQuery,
			cellID, timestamp, yCount, lag1hCount, lag24hCount, lag7dCount,
			dow, hour, isHoliday, isWeekend, isBusinessHours)

		if err != nil {
			kg.logger.Printf("⚠️  Erro ao inserir feature: %v", err)
		}
	}

	return nil
}

// ============================================================================
// FASE 5: VALIDAÇÃO DE QUALIDADE
// ============================================================================

func (kg *KnowledgeBaseGenerator) validateDataQuality(ctx context.Context, db *sql.DB) error {
	metrics := make(map[string]interface{})

	// Total de incidentes
	var incidentsCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM curated_incidents`).Scan(&incidentsCount); err != nil {
		kg.logger.Printf("⚠️  Erro ao contar incidentes: %v", err)
	}
	metrics["total_incidents"] = incidentsCount

	// Total de células
	var cellsCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM curated_cells`).Scan(&cellsCount); err != nil {
		kg.logger.Printf("⚠️  Erro ao contar células: %v", err)
	}
	metrics["total_cells"] = cellsCount

	// Total de features
	var featuresCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM features_cell_hourly`).Scan(&featuresCount); err != nil {
		kg.logger.Printf("⚠️  Erro ao contar features: %v", err)
	}
	metrics["total_features"] = featuresCount

	// Persistir métricas
	metricsJSON, _ := json.Marshal(metrics)
	_, err := db.ExecContext(ctx, `
		INSERT INTO analytics_quality_reports (report_date, metrics)
		VALUES (CURDATE(), ?)
		ON DUPLICATE KEY UPDATE metrics = VALUES(metrics)
	`, string(metricsJSON))

	kg.logger.Printf("📊 Métricas de qualidade:")
	kg.logger.Printf("   • Total de incidentes: %d", incidentsCount)
	kg.logger.Printf("   • Total de células: %d", cellsCount)
	kg.logger.Printf("   • Total de features: %d", featuresCount)

	return err
}
