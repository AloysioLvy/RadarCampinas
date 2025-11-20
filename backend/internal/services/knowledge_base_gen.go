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
// FASE 1: MIGRAÇÃO DE DADOS HISTÓRICOS (COM BATCH INSERT)
// ============================================================================

func (kg *KnowledgeBaseGenerator) migrateHistoricalData(ctx context.Context, db *sql.DB) error {
	query := `
		SELECT r.report_id, r.neighborhood_id, r.crime_id, r.report_date, r.created_at, r.updated_at,
			n.name as neighborhood_name, n.latitude, n.longitude, n.neighborhood_weight,
			c.crime_name, c.crime_weight
		FROM reports r
		JOIN neighborhoods n ON r.neighborhood_id = n.neighborhood_id
		JOIN crimes c ON r.crime_id = c.crime_id
		WHERE r.report_date BETWEEN @p1 AND @p2
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
	batchSize := kg.config.BatchSize
	if batchSize <= 0 {
		batchSize = 500
	}

	var batch []Report

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
		batch = append(batch, report)

		if len(batch) >= batchSize {
			ok, fail := kg.insertIncidentsBatch(ctx, db, batch)
			processed += ok
			skipped += fail
			batch = batch[:0]
			kg.logger.Printf("  ➜ Processados %d registros (batch)...", processed)
		}
	}

	if len(batch) > 0 {
		ok, fail := kg.insertIncidentsBatch(ctx, db, batch)
		processed += ok
		skipped += fail
	}

	kg.logger.Printf("✅ Migração concluída: %d incidentes processados, %d ignorados", processed, skipped)
	return nil
}

// insertIncidentsBatch faz insert em lote de curated_incidents
func (kg *KnowledgeBaseGenerator) insertIncidentsBatch(ctx context.Context, db *sql.DB, reports []Report) (processed int, skipped int) {
	if len(reports) == 0 {
		return 0, 0
	}

	valueStrings := []string{}
	valueArgs := []interface{}{}
	paramIndex := 1

	for _, report := range reports {
		lat, err := strconv.ParseFloat(report.Neighborhood.Latitude, 64)
		if err != nil {
			skipped++
			continue
		}
		lon, err := strconv.ParseFloat(report.Neighborhood.Longitude, 64)
		if err != nil {
			skipped++
			continue
		}

		if lat < -23.1 || lat > -22.7 || lon < -47.3 || lon > -46.8 {
			skipped++
			continue
		}

		reportTime, err := time.Parse("2006-01-02", report.ReportDate)
		if err != nil {
			if reportTime, err = time.Parse("2006-01-02 15:04:05", report.ReportDate); err != nil {
				skipped++
				continue
			}
		}

		category := kg.mapCrimeCategory(report.Crime.CrimeName)
		severity := report.Crime.CrimeWeight
		confidence := kg.calculateConfidence(report)
		incidentID := fmt.Sprintf("rpt_%d", report.ReportID)

		valueStrings = append(valueStrings, fmt.Sprintf(
			"(@p%d, @p%d, @p%d, @p%d, @p%d, @p%d, @p%d, @p%d, @p%d)",
			paramIndex, paramIndex+1, paramIndex+2, paramIndex+3,
			paramIndex+4, paramIndex+5, paramIndex+6, paramIndex+7, paramIndex+8,
		))

		valueArgs = append(valueArgs,
			incidentID,
			reportTime,
			category,
			severity,
			lat,
			lon,
			report.Neighborhood.Name,
			confidence,
			"legacy_reports",
		)

		paramIndex += 9
		processed++
	}

	if len(valueStrings) == 0 {
		return 0, skipped
	}

	query := fmt.Sprintf(`
		INSERT INTO curated_incidents 
		(id, occurred_at, category, severity, latitude, longitude, neighborhood, confidence, source)
		VALUES %s
	`, strings.Join(valueStrings, ","))

	if _, err := db.ExecContext(ctx, query, valueArgs...); err != nil {
		kg.logger.Printf("⚠️  Erro no batch insert de incidents: %v", err)
		// Se der erro, conta todos do batch como ignorados
		skipped += processed
		processed = 0
	}

	return processed, skipped
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

			cellIDStr := fmt.Sprintf("CAMP-%d-%d", kg.config.CellResolution, cellID)

			insertQuery := `
				IF NOT EXISTS (SELECT 1 FROM curated_cells WHERE cell_id = @p1)
				BEGIN
					INSERT INTO curated_cells 
					(cell_id, cell_resolution, city, center_lat, center_lng)
					VALUES (@p2, @p3, @p4, @p5, @p6)
				END
			`

			_, err := db.ExecContext(ctx, insertQuery,
				cellIDStr,
				cellIDStr,
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
	cellsQuery := `SELECT cell_id, center_lat, center_lng FROM curated_cells WHERE cell_resolution = @p1`
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

	var cells []Cell
	for cellRows.Next() {
		var c Cell
		if err := cellRows.Scan(&c.ID, &c.CenterLat, &c.CenterLng); err != nil {
			continue
		}
		cells = append(cells, c)
	}

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

		for _, cell := range cells {
			if lat >= cell.CenterLat-cellSizeDegrees/2 && lat < cell.CenterLat+cellSizeDegrees/2 &&
				lon >= cell.CenterLng-cellSizeDegrees/2 && lon < cell.CenterLng+cellSizeDegrees/2 {

				updateQuery := `UPDATE curated_incidents SET cell_id = @p1, cell_resolution = @p2 WHERE id = @p3`
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
func (kg *KnowledgeBaseGenerator) generateFeaturesForRange(ctx context.Context, db *sql.DB, rangeStart, rangeEnd time.Time) error {
	query := `
	;WITH Hours AS (
		-- Gera todas as horas entre @start e @end
		SELECT DATEADD(hour, n, @start) AS ts
		FROM (
			SELECT TOP (DATEDIFF(hour, @start, @end))
				ROW_NUMBER() OVER (ORDER BY (SELECT NULL)) - 1 AS n
			FROM sys.all_objects
		) AS t
	),
	Cells AS (
		SELECT cell_id, center_lat, center_lng
		FROM curated_cells
		WHERE cell_resolution = @cellRes
	),
	CellHours AS (
		-- Produto cartesiano de células x horas
		SELECT c.cell_id, h.ts
		FROM Cells c
		CROSS JOIN Hours h
	),
	BaseIncidents AS (
		-- Incidentes no range estendido para calcular lags
		SELECT 
			cell_id,
			occurred_at
		FROM curated_incidents
		WHERE occurred_at >= DATEADD(day, -7, @start)
		  AND occurred_at < @end
	),
	Aggregated AS (
		SELECT
			ch.cell_id,
			ch.ts,

			-- y_count: ocorrências na própria hora
			SUM(CASE 
					WHEN bi.occurred_at >= ch.ts
					 AND bi.occurred_at < DATEADD(hour, 1, ch.ts)
					THEN 1 ELSE 0 END
			) AS y_count,

			-- lag_1h: ocorrências na hora anterior
			SUM(CASE 
					WHEN bi.occurred_at >= DATEADD(hour, -1, ch.ts)
					 AND bi.occurred_at < ch.ts
					THEN 1 ELSE 0 END
			) AS lag_1h,

			-- lag_24h: últimas 24h
			SUM(CASE 
					WHEN bi.occurred_at >= DATEADD(hour, -24, ch.ts)
					 AND bi.occurred_at < ch.ts
					THEN 1 ELSE 0 END
			) AS lag_24h,

			-- lag_7d: últimos 7 dias
			SUM(CASE 
					WHEN bi.occurred_at >= DATEADD(day, -7, ch.ts)
					 AND bi.occurred_at < ch.ts
					THEN 1 ELSE 0 END
			) AS lag_7d
		FROM CellHours ch
		LEFT JOIN BaseIncidents bi
			ON bi.cell_id = ch.cell_id
		   AND bi.occurred_at >= DATEADD(day, -7, ch.ts)
		   AND bi.occurred_at < DATEADD(hour, 1, ch.ts)
		GROUP BY ch.cell_id, ch.ts
	),
	WithCalendar AS (
		SELECT
			a.cell_id,
			a.ts,
			a.y_count,
			a.lag_1h,
			a.lag_24h,
			a.lag_7d,
			DATEPART(dw, a.ts) - 1 AS dow, -- 0=domingo
			DATEPART(hour, a.ts) AS hour,
			CASE WHEN eh.date IS NULL THEN 0 ELSE 1 END AS holiday,
			CASE WHEN DATEPART(dw, a.ts) IN (1, 7) THEN 1 ELSE 0 END AS is_weekend,
			CASE WHEN DATEPART(hour, a.ts) BETWEEN 8 AND 18 THEN 1 ELSE 0 END AS is_business_hours
		FROM Aggregated a
		LEFT JOIN external_holidays eh
			ON eh.date = CAST(a.ts AS date)
	)
	-- INSERT ou UPDATE em features_cell_hourly em batch
	MERGE features_cell_hourly AS target
	USING WithCalendar AS source
		ON target.cell_id = source.cell_id
	   AND target.ts = source.ts
	WHEN MATCHED THEN
		UPDATE SET
			y_count = source.y_count,
			lag_1h = source.lag_1h,
			lag_24h = source.lag_24h,
			lag_7d = source.lag_7d,
			dow = source.dow,
			hour = source.hour,
			holiday = source.holiday,
			is_weekend = source.is_weekend,
			is_business_hours = source.is_business_hours
	WHEN NOT MATCHED BY TARGET THEN
		INSERT (cell_id, ts, y_count, lag_1h, lag_24h, lag_7d, dow, hour, holiday, is_weekend, is_business_hours)
		VALUES (source.cell_id, source.ts, source.y_count, source.lag_1h, source.lag_24h, source.lag_7d,
				source.dow, source.hour, source.holiday, source.is_weekend, source.is_business_hours);
	`

	_, err := db.ExecContext(ctx, query,
		sql.Named("start", rangeStart),
		sql.Named("end", rangeEnd),
		sql.Named("cellRes", kg.config.CellResolution),
	)
	if err != nil {
		return fmt.Errorf("erro ao gerar features para range %s - %s: %w",
			rangeStart.Format(time.RFC3339), rangeEnd.Format(time.RFC3339), err)
	}

	return nil
}
func (kg *KnowledgeBaseGenerator) generateTemporalFeatures(ctx context.Context, db *sql.DB) error {
	start := kg.config.StartDate
	end := kg.config.EndDate

	// Processar em blocos de 1 dia
	day := start
	daysProcessed := 0

	for day.Before(end) {
		dayEnd := day.Add(24 * time.Hour)
		if dayEnd.After(end) {
			dayEnd = end
		}

		startBlock := time.Now()
		if err := kg.generateFeaturesForRange(ctx, db, day, dayEnd); err != nil {
			kg.logger.Printf("⚠️  Erro ao processar dia %s: %v", day.Format("2006-01-02"), err)
		}

		daysProcessed++
		kg.logger.Printf("  ➜ Dia %s processado em %s (total dias: %d)",
			day.Format("2006-01-02"), time.Since(startBlock), daysProcessed)

		day = dayEnd
	}

	kg.logger.Printf("✅ Features temporais geradas para %d dias", daysProcessed)
	return nil
}


func (kg *KnowledgeBaseGenerator) generateHourlyFeatures(ctx context.Context, db *sql.DB, timestamp time.Time) error {
	cellsQuery := `SELECT cell_id FROM curated_cells WHERE cell_resolution = @p1`
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

		var yCount int
		countQuery := `
			SELECT COUNT(*) FROM curated_incidents 
			WHERE cell_id = @p1 AND occurred_at >= @p2 AND occurred_at < @p3
		`
		if err := db.QueryRowContext(ctx, countQuery, cellID, timestamp, endHour).Scan(&yCount); err != nil {
			kg.logger.Printf("⚠️  Erro ao contar crimes: %v", err)
			continue
		}

		lag1h := timestamp.Add(-time.Hour)
		lag24h := timestamp.Add(-24 * time.Hour)
		lag7d := timestamp.Add(-7 * 24 * time.Hour)

		var lag1hCount, lag24hCount, lag7dCount int

		if err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM curated_incidents WHERE cell_id = @p1 AND occurred_at >= @p2 AND occurred_at < @p3`,
			cellID, lag1h, timestamp).Scan(&lag1hCount); err != nil {
			kg.logger.Printf("⚠️  Erro ao calcular lag 1h: %v", err)
		}
		if err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM curated_incidents WHERE cell_id = @p1 AND occurred_at >= @p2 AND occurred_at < @p3`,
			cellID, lag24h, timestamp).Scan(&lag24hCount); err != nil {
			kg.logger.Printf("⚠️  Erro ao calcular lag 24h: %v", err)
		}
		if err := db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM curated_incidents WHERE cell_id = @p1 AND occurred_at >= @p2 AND occurred_at < @p3`,
			cellID, lag7d, timestamp).Scan(&lag7dCount); err != nil {
			kg.logger.Printf("⚠️  Erro ao calcular lag 7d: %v", err)
		}

		var isHoliday bool
		if err := db.QueryRowContext(ctx,
			`SELECT CASE WHEN COUNT(*) > 0 THEN 1 ELSE 0 END FROM external_holidays WHERE date = @p1`,
			timestamp.Format("2006-01-02")).Scan(&isHoliday); err != nil {
			kg.logger.Printf("⚠️  Erro ao verificar feriado: %v", err)
		}

		dow := int(timestamp.Weekday())
		hour := timestamp.Hour()
		isWeekend := dow == 0 || dow == 6
		isBusinessHours := hour >= 8 && hour <= 18

		insertQuery := `
			INSERT INTO features_cell_hourly 
			(cell_id, ts, y_count, lag_1h, lag_24h, lag_7d, dow, hour, holiday, is_weekend, is_business_hours)
			VALUES (@p1, @p2, @p3, @p4, @p5, @p6, @p7, @p8, @p9, @p10, @p11)
		`

		_, err := db.ExecContext(ctx, insertQuery,
			cellID, timestamp, yCount, lag1hCount, lag24hCount, lag7dCount,
			dow, hour, isHoliday, isWeekend, isBusinessHours)

		if err != nil && strings.Contains(err.Error(), "2627") {
			updateQuery := `
				UPDATE features_cell_hourly 
				SET y_count = @p1, lag_1h = @p2, lag_24h = @p3, lag_7d = @p4
				WHERE cell_id = @p5 AND ts = @p6
			`
			_, err = db.ExecContext(ctx, updateQuery,
				yCount, lag1hCount, lag24hCount, lag7dCount, cellID, timestamp)
		}

		if err != nil {
			kg.logger.Printf("⚠️  Erro ao inserir/atualizar feature: %v", err)
		}
	}

	return nil
}

// ============================================================================
// FASE 5: VALIDAÇÃO DE QUALIDADE
// ============================================================================

func (kg *KnowledgeBaseGenerator) validateDataQuality(ctx context.Context, db *sql.DB) error {
	metrics := make(map[string]interface{})

	var incidentsCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM curated_incidents`).Scan(&incidentsCount); err != nil {
		kg.logger.Printf("⚠️  Erro ao contar incidentes: %v", err)
	}
	metrics["total_incidents"] = incidentsCount

	var cellsCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM curated_cells`).Scan(&cellsCount); err != nil {
		kg.logger.Printf("⚠️  Erro ao contar células: %v", err)
	}
	metrics["total_cells"] = cellsCount

	var featuresCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM features_cell_hourly`).Scan(&featuresCount); err != nil {
		kg.logger.Printf("⚠️  Erro ao contar features: %v", err)
	}
	metrics["total_features"] = featuresCount

	metricsJSON, _ := json.Marshal(metrics)

	insertQuery := `
		INSERT INTO analytics_quality_reports (report_date, metrics)
		VALUES (CAST(GETDATE() AS DATE), @p1)
	`
	_, err := db.ExecContext(ctx, insertQuery, string(metricsJSON))

	if err != nil && strings.Contains(err.Error(), "2627") {
		updateQuery := `
			UPDATE analytics_quality_reports 
			SET metrics = @p1
			WHERE report_date = CAST(GETDATE() AS DATE)
		`
		_, err = db.ExecContext(ctx, updateQuery, string(metricsJSON))
	}

	kg.logger.Printf("📊 Métricas de qualidade:")
	kg.logger.Printf("   • Total de incidentes: %d", incidentsCount)
	kg.logger.Printf("   • Total de células: %d", cellsCount)
	kg.logger.Printf("   • Total de features: %d", featuresCount)

	return err
}