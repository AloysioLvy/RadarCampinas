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

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/database/migrations"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
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

type Holiday struct {
	Date time.Time `json:"date"`
	Name string    `json:"name"`
	Type string    `json:"type"`
}

type KnowledgeBaseConfig struct {
	SourceDB       *sql.DB
	TargetDB       *pgxpool.Pool
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
		executionID: uuid.New().String(),
	}
}

// ============================================================================
// MIGRATIONS: Executar migrações SQL automaticamente
// ============================================================================

// runMigrations executa as migrations SQL para criar schemas e tabelas
// Esta função é idempotente - pode ser executada múltiplas vezes com segurança
func (kg *KnowledgeBaseGenerator) runMigrations(ctx context.Context) error {
	kg.logger.Println("🔧 Verificando e aplicando migrations...")
	kg.logPhase(ctx, "migrate", "running", 0)

	// Ler arquivo de migrations embutido
	migrationsSQL, err := migrations.Files.ReadFile("knowledge_base_schema.sql")
	if err != nil {
		// Fallback: tentar ler do filesystem para ambiente de desenvolvimento
		kg.logger.Println("⚠️  Não consegui ler migrations embutidas, tentando do filesystem...")
		migrationsSQL, err = os.ReadFile("backend/internal/database/migrations/knowledge_base_schema.sql")
		if err != nil {
			// segundo fallback: caminho relativo quando o binário roda a partir da raiz do módulo backend/
			migrationsSQL, err = os.ReadFile("internal/database/migrations/knowledge_base_schema.sql")
			if err != nil {
				return fmt.Errorf("erro ao ler migrations: %v", err)
			}
		}
	}

	// Verificar se migrations já foram aplicadas
	var migrationExists bool
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM pg_tables 
			WHERE schemaname = 'public' AND tablename = 'schema_migrations'
		)`

	err = kg.config.TargetDB.QueryRow(ctx, checkQuery).Scan(&migrationExists)
	if err != nil {
		kg.logger.Printf("⚠️  Erro ao verificar migrations existentes: %v", err)
	}

	if migrationExists {
		var versionExists bool
		versionQuery := `
			SELECT EXISTS (
				SELECT 1 FROM public.schema_migrations 
				WHERE version = 'knowledge_base_v1.0.0'
			)`
		err = kg.config.TargetDB.QueryRow(ctx, versionQuery).Scan(&versionExists)
		if err == nil && versionExists {
			kg.logger.Println("✅ Migrations já aplicadas anteriormente (idempotente)")
			kg.logPhase(ctx, "migrate", "success", 0)
			return nil
		}
	}

	// Executar migrations
	kg.logger.Println("📦 Aplicando migrations SQL...")
	startTime := time.Now()

	_, err = kg.config.TargetDB.Exec(ctx, string(migrationsSQL))
	if err != nil {
		kg.logPhase(ctx, "migrate", "failed", 0)
		return fmt.Errorf("erro ao executar migrations: %v", err)
	}

	executionTime := int(time.Since(startTime).Seconds())
	kg.logger.Printf("✅ Migrations aplicadas com sucesso em %ds", executionTime)

	// Verificar schemas criados
	kg.verifySchemas(ctx)

	kg.logPhase(ctx, "migrate", "success", 0)
	return nil
}

// verifySchemas verifica se todos os schemas foram criados corretamente
func (kg *KnowledgeBaseGenerator) verifySchemas(ctx context.Context) {
	expectedSchemas := []string{"curated", "external", "features", "analytics"}

	for _, schema := range expectedSchemas {
		var exists bool
		query := `SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = $1)`
		err := kg.config.TargetDB.QueryRow(ctx, query, schema).Scan(&exists)

		if err != nil || !exists {
			kg.logger.Printf("❌ Schema '%s' não encontrado!", schema)
		} else {
			kg.logger.Printf("✓ Schema '%s' verificado", schema)
		}
	}
}

// logPhase registra uma fase do pipeline no banco de dados
func (kg *KnowledgeBaseGenerator) logPhase(ctx context.Context, phase, status string, recordsProcessed int) {
	query := `
		INSERT INTO analytics.pipeline_logs (execution_id, started_at, status, phase, records_processed)
		VALUES ($1, NOW(), $2, $3, $4)
		ON CONFLICT (execution_id) DO UPDATE SET
			status = EXCLUDED.status,
			finished_at = CASE WHEN EXCLUDED.status IN ('success', 'failed') THEN NOW() ELSE NULL END,
			records_processed = EXCLUDED.records_processed
	`

	_, err := kg.config.TargetDB.Exec(ctx, query, kg.executionID, status, phase, recordsProcessed)
	if err != nil {
		kg.logger.Printf("⚠️  Erro ao registrar log: %v", err)
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

	// Fase 0: Executar migrations (NOVO!)
	if err := kg.runMigrations(ctx); err != nil {
		return fmt.Errorf("❌ erro nas migrations: %v", err)
	}

	// Fase 1: Migrar dados históricos
	if err := kg.migrateHistoricalData(ctx); err != nil {
		kg.logPhase(ctx, "migrate_data", "failed", 0)
		return fmt.Errorf("❌ erro na migração de dados históricos: %v", err)
	}

	// Fase 2: Gerar grade de células
	if err := kg.generateSpatialGrid(ctx); err != nil {
		kg.logPhase(ctx, "spatial_grid", "failed", 0)
		return fmt.Errorf("❌ erro na geração da grade espacial: %v", err)
	}

	// Fase 3: Atribuir células aos incidentes
	if err := kg.assignCellsToIncidents(ctx); err != nil {
		kg.logPhase(ctx, "assign_cells", "failed", 0)
		return fmt.Errorf("❌ erro na atribuição de células: %v", err)
	}

	// Fase 4: Ingerir dados externos
	if err := kg.ingestExternalData(ctx); err != nil {
		kg.logPhase(ctx, "external_data", "failed", 0)
		return fmt.Errorf("❌ erro na ingestão de dados externos: %v", err)
	}

	// Fase 5: Gerar features temporais
	if err := kg.generateTemporalFeatures(ctx); err != nil {
		kg.logPhase(ctx, "features", "failed", 0)
		return fmt.Errorf("❌ erro na geração de features: %v", err)
	}

	// Fase 6: Validar qualidade dos dados
	if err := kg.validateDataQuality(ctx); err != nil {
		kg.logPhase(ctx, "validation", "failed", 0)
		return fmt.Errorf("❌ erro na validação de qualidade: %v", err)
	}

	executionTime := time.Since(startTime)
	kg.logger.Printf("✅ Base de conhecimento gerada com sucesso em %s!", executionTime)
	kg.logPhase(ctx, "complete", "success", 0)

	return nil
}

// ============================================================================
// IMPLEMENTAÇÃO DAS FASES (código original com melhorias de logging)
// ============================================================================

func (kg *KnowledgeBaseGenerator) migrateHistoricalData(ctx context.Context) error {
	kg.logger.Println("📊 Fase 1: Migrando dados históricos...")
	kg.logPhase(ctx, "migrate_data", "running", 0)

	query := `
		SELECT r.report_id, r.neighborhood_id, r.crime_id, r.report_date, r.created_at, r.updated_at,
			n.name as neighborhood_name, n.latitude, n.longitude, n.neighborhood_weight,
			c.crime_name, c.crime_weight
		FROM reports r
		JOIN neighborhoods n ON r.neighborhood_id = n.neighborhood_id
		JOIN crimes c ON r.crime_id = c.crime_id
		WHERE r.report_date BETWEEN $1 AND $2
		ORDER BY r.report_date
	`

	rows, err := kg.config.SourceDB.QueryContext(ctx, query,
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

	batch := make([]map[string]interface{}, 0, kg.config.BatchSize)
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

		incident, err := kg.mapReportToIncident(report)
		if err != nil {
			kg.logger.Printf("⚠️  Erro ao mapear report %d: %v", report.ReportID, err)
			skipped++
			continue
		}

		batch = append(batch, incident)

		if len(batch) >= kg.config.BatchSize {
			if err := kg.insertIncidentsBatch(ctx, batch); err != nil {
				return fmt.Errorf("erro ao inserir batch: %v", err)
			}
			processed += len(batch)
			batch = batch[:0]
			kg.logger.Printf("  ➜ Processados %d registros (%d ignorados)...", processed, skipped)
		}
	}

	if len(batch) > 0 {
		if err := kg.insertIncidentsBatch(ctx, batch); err != nil {
			return fmt.Errorf("erro ao inserir último batch: %v", err)
		}
		processed += len(batch)
	}

	kg.logger.Printf("✅ Migração concluída: %d incidentes processados, %d ignorados", processed, skipped)
	kg.logPhase(ctx, "migrate_data", "success", processed)
	return nil
}

func (kg *KnowledgeBaseGenerator) mapReportToIncident(report Report) (map[string]interface{}, error) {
	lat, err := strconv.ParseFloat(report.Neighborhood.Latitude, 64)
	if err != nil {
		return nil, fmt.Errorf("latitude inválida: %v", err)
	}

	lon, err := strconv.ParseFloat(report.Neighborhood.Longitude, 64)
	if err != nil {
		return nil, fmt.Errorf("longitude inválida: %v", err)
	}

	if !kg.isWithinCampinas(lat, lon) {
		return nil, fmt.Errorf("coordenadas fora de Campinas")
	}

	reportTime, err := time.Parse("2006-01-02", report.ReportDate)
	if err != nil {
		if reportTime, err = time.Parse("2006-01-02 15:04:05", report.ReportDate); err != nil {
			if reportTime, err = time.Parse("2006-01-02T15:04:05", report.ReportDate); err != nil {
				return nil, fmt.Errorf("formato de data inválido: %v", err)
			}
		}
	}

	category := kg.mapCrimeCategory(report.Crime.CrimeName)
	severity := report.Crime.CrimeWeight

	incident := map[string]interface{}{
		"id":           fmt.Sprintf("rpt_%d", report.ReportID),
		"occurred_at":  reportTime,
		"category":     category,
		"severity":     severity,
		"lat":          lat,
		"lon":          lon,
		"neighborhood": report.Neighborhood.Name,
		"confidence":   kg.calculateConfidence(report),
		"source":       "legacy_reports",
	}

	return incident, nil
}

func (kg *KnowledgeBaseGenerator) isWithinCampinas(lat, lon float64) bool {
	minLat, maxLat := -23.1, -22.7
	minLon, maxLon := -47.3, -46.8
	return lat >= minLat && lat <= maxLat && lon >= minLon && lon <= maxLon
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

func (kg *KnowledgeBaseGenerator) insertIncidentsBatch(ctx context.Context, incidents []map[string]interface{}) error {
	if len(incidents) == 0 {
		return nil
	}

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
	kg.logger.Printf("🗺️  Fase 2: Gerando grade espacial de %dm...", kg.config.CellResolution)
	kg.logPhase(ctx, "spatial_grid", "running", 0)

	minLon, minLat := -47.3, -23.1
	maxLon, maxLat := -46.8, -22.7

	cellSizeDegrees := float64(kg.config.CellResolution) / 111000.0

	cells := make([]map[string]interface{}, 0)
	cellID := 0

	for lon := minLon; lon < maxLon; lon += cellSizeDegrees {
		for lat := minLat; lat < maxLat; lat += cellSizeDegrees {
			cellID++

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

	err := kg.insertCellsBatch(ctx, cells)
	if err != nil {
		return err
	}

	kg.logger.Printf("✅ Grade espacial gerada: %d células", len(cells))
	kg.logPhase(ctx, "spatial_grid", "success", len(cells))
	return nil
}

func (kg *KnowledgeBaseGenerator) insertCellsBatch(ctx context.Context, cells []map[string]interface{}) error {
	if len(cells) == 0 {
		return nil
	}

	const batchSize = 1000 // Número seguro para evitar limite de parâmetros

	for start := 0; start < len(cells); start += batchSize {
		end := start + batchSize
		if end > len(cells) {
			end = len(cells)
		}

		batch := cells[start:end]

		valueStrings := make([]string, 0, len(batch))
		valueArgs := make([]interface{}, 0, len(batch)*7)

		for i, cell := range batch {
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
	}

	return nil
}

func (kg *KnowledgeBaseGenerator) assignCellsToIncidents(ctx context.Context) error {
	kg.logger.Println("🎯 Fase 3: Atribuindo células aos incidentes...")
	kg.logPhase(ctx, "assign_cells", "running", 0)

	query := `
		UPDATE curated.incidents 
		SET cell_id = c.cell_id, cell_resolution = c.cell_resolution
		FROM curated.cells c
		WHERE curated.incidents.cell_id IS NULL
		  AND c.cell_resolution = $1
		  AND ST_Contains(c.geom::geometry, curated.incidents.geom::geometry)
	`

	result, err := kg.config.TargetDB.Exec(ctx, query, kg.config.CellResolution)
	if err != nil {
		return err
	}

	rowsAffected := result.RowsAffected()
	kg.logger.Printf("✅ Atribuídas células a %d incidentes", rowsAffected)
	kg.logPhase(ctx, "assign_cells", "success", int(rowsAffected))

	return nil
}

func (kg *KnowledgeBaseGenerator) ingestExternalData(ctx context.Context) error {
	kg.logger.Println("📅 Fase 4: Verificando feriados...")
	kg.logPhase(ctx, "external_data", "running", 0)

	// Apenas verificar se feriados existem (já foram criados na migration)
	var holidayCount int
	err := kg.config.TargetDB.QueryRow(ctx, "SELECT COUNT(*) FROM external.holidays").Scan(&holidayCount)
	if err != nil {
		return err
	}

	kg.logger.Printf("✅ Feriados disponíveis: %d registros", holidayCount)
	kg.logPhase(ctx, "external_data", "success", holidayCount)
	return nil
}

func (kg *KnowledgeBaseGenerator) generateTemporalFeatures(ctx context.Context) error {
	kg.logger.Println("⚙️  Fase 5: Gerando features temporais...")
	kg.logPhase(ctx, "features", "running", 0)

	current := kg.config.StartDate
	hoursProcessed := 0

	for current.Before(kg.config.EndDate) {
		if err := kg.generateHourlyFeatures(ctx, current); err != nil {
			return err
		}
		hoursProcessed++
		current = current.Add(time.Hour)

		if hoursProcessed%100 == 0 {
			kg.logger.Printf("  ➜ Processadas %d horas...", hoursProcessed)
		}
	}

	kg.logger.Printf("✅ Features temporais geradas: %d horas processadas", hoursProcessed)
	kg.logPhase(ctx, "features", "success", hoursProcessed)
	return nil
}

func (kg *KnowledgeBaseGenerator) generateHourlyFeatures(ctx context.Context, timestamp time.Time) error {
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
		holiday_check AS (
			SELECT COUNT(*) > 0 as holiday
			FROM external.holidays
			WHERE date = $1::date
		)
		INSERT INTO features.cell_hourly (
			cell_id, ts, y_count, lag_1h, lag_24h, lag_7d,
			roll_3h_sum, roll_24h_sum, roll_7d_sum,
			dow, hour, holiday, is_weekend, is_business_hours
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
			hc_check.holiday,
			EXTRACT(DOW FROM $1) IN (0, 6) as is_weekend,
			EXTRACT(HOUR FROM $1) BETWEEN 8 AND 18 as is_business_hours
		FROM curated.cells c
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
			holiday = EXCLUDED.holiday,
			is_weekend = EXCLUDED.is_weekend,
			is_business_hours = EXCLUDED.is_business_hours
	`

	endHour := timestamp.Add(time.Hour)
	lag1h := timestamp.Add(-time.Hour)
	lag24h := timestamp.Add(-24 * time.Hour)
	lag7d := timestamp.Add(-7 * 24 * time.Hour)
	roll3h := timestamp.Add(-3 * time.Hour)

	_, err := kg.config.TargetDB.Exec(ctx, query,
		timestamp, endHour, kg.config.CellResolution,
		lag1h, lag24h, lag7d, roll3h)

	return err
}

func (kg *KnowledgeBaseGenerator) validateDataQuality(ctx context.Context) error {
	kg.logger.Println("✓ Fase 6: Validando qualidade dos dados...")
	kg.logPhase(ctx, "validation", "running", 0)

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

	spatialCoverage := 0.0
	if totalCells > 0 {
		spatialCoverage = float64(cellsWithData) / float64(totalCells)
	}
	metrics["spatial_coverage"] = spatialCoverage

	// 2. Contagem de incidentes
	var incidentsCount int
	if err := kg.config.TargetDB.QueryRow(ctx, `SELECT COUNT(*) FROM curated.incidents`).Scan(&incidentsCount); err != nil {
		return err
	}
	metrics["total_incidents"] = incidentsCount

	// 3. Cobertura temporal
	var temporalCoverage float64
	if incidentsCount > 0 {
		var minDate, maxDate time.Time
		var hoursWithData int
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
		totalHours := int(maxDate.Sub(minDate).Hours())
		if totalHours > 0 {
			temporalCoverage = float64(hoursWithData) / float64(totalHours)
		}
	}
	metrics["temporal_coverage"] = temporalCoverage

	// 4. Taxa de duplicação
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

	duplicationRate := 0.0
	if totalReports > 0 {
		duplicationRate = 1.0 - (float64(uniqueReports) / float64(totalReports))
	}
	metrics["duplication_rate"] = duplicationRate

	// 5. Persistir métricas
	metricsJSON, _ := json.Marshal(metrics)
	_, err = kg.config.TargetDB.Exec(ctx, `
		INSERT INTO analytics.quality_reports (report_date, metrics, created_at)
		VALUES (CURRENT_DATE, $1, NOW())
		ON CONFLICT (report_date) DO UPDATE SET
			metrics = EXCLUDED.metrics,
			updated_at = NOW()
	`, string(metricsJSON))
	if err != nil {
		return err
	}

	// 6. Log das métricas
	kg.logger.Printf("📊 Métricas de qualidade:")
	kg.logger.Printf("   • Total de incidentes: %d", incidentsCount)
	kg.logger.Printf("   • Cobertura espacial: %.2f%%", spatialCoverage*100)
	kg.logger.Printf("   • Cobertura temporal: %.2f%%", temporalCoverage*100)
	kg.logger.Printf("   • Taxa de duplicação: %.2f%%", duplicationRate*100)

	kg.logger.Println("✅ Validação de qualidade concluída")
	kg.logPhase(ctx, "validation", "success", 0)
	return nil
}
