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
	ReportDateFormated string	`json:"report_date_formated"`
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

	// Fase 2.5: Mapeamento célula → bairro
	kg.logger.Println("🏷️ Fase 2.5: Gerando mapeamento célula → bairro...")
	if err := kg.mapCellsToNeighborhoods(ctx, db); err != nil {
		return fmt.Errorf("erro no mapeamento de células para bairros: %v", err)
	}

	// Fase 3: Atribuir células aos incidentes
	kg.logger.Println("🎯 Fase 3: Atribuindo células aos incidentes...")
	if err := kg.assignCellsToIncidents(ctx, db); err != nil {
		return fmt.Errorf("❌ erro na atribuição de células: %v", err)
	}

	// Fase 3.5: Gerar features mensais
	kg.logger.Println("📅 Fase 3.5: Gerando features mensais...")
	if err := kg.generateMonthlyFeatures(ctx, db); err != nil {
		return fmt.Errorf("erro na geração de features mensais: %v", err)
	}

	// Fase 4: Gerar features temporais (horárias)
	////kg.logger.Println("⚙️  Fase 4: Gerando features temporais...")
	////if err := kg.generateTemporalFeatures(ctx, db); err != nil {
	////	return fmt.Errorf("❌ erro na geração de features: %v", err)
	////}

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
    SELECT r.report_id, r.neighborhood_id, r.crime_id, r.report_date_formated, r.created_at, r.updated_at,
        n.name as neighborhood_name, n.latitude, n.longitude, n.neighborhood_weight,
        c.crime_name, c.crime_weight
    FROM reports r
    JOIN neighborhoods n ON r.neighborhood_id = n.neighborhood_id
    JOIN crimes c ON r.crime_id = c.crime_id
    WHERE TRY_CONVERT(date, r.report_date_formated, 103) BETWEEN @start AND @end
    ORDER BY TRY_CONVERT(date, r.report_date_formated, 103)
`

	rows, err := db.QueryContext(ctx, query,
    sql.Named("start", kg.config.StartDate),
    sql.Named("end", kg.config.EndDate),
		)
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
			&report.ReportDateFormated, &report.CreatedAt, &report.UpdatedAt,
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

func (kg *KnowledgeBaseGenerator) insertIncidentsBatch(ctx context.Context, db *sql.DB, reports []Report) (processed int, skipped int) {
	if len(reports) == 0 {
		return 0, 0
	}

	const maxParams = 2100
	const paramsPerRow = 9
	maxRowsPerInsert := maxParams / paramsPerRow // 233

	valueStrings := []string{}
	valueArgs := []interface{}{}
	paramIndex := 1

	flushBatch := func() {
		if len(valueStrings) == 0 {
			return
		}
		query := fmt.Sprintf(`
            INSERT INTO curated_incidents 
            (id, occurred_at, category, severity, latitude, longitude, neighborhood, confidence, source)
            VALUES %s
        `, strings.Join(valueStrings, ","))

		if _, err := db.ExecContext(ctx, query, valueArgs...); err != nil {
			kg.logger.Printf("⚠️  Erro no batch insert de incidents: %v", err)
			// se der erro nesse sub-batch, considera todos eles como ignorados
			skipped += len(valueStrings)
		} else {
			processed += len(valueStrings)
		}

		// reset do batch
		valueStrings = valueStrings[:0]
		valueArgs = valueArgs[:0]
		paramIndex = 1
	}

	for _, report := range reports {
		lat, err := strconv.ParseFloat(strings.ReplaceAll(report.Neighborhood.Latitude, ",", "."), 64)
		if err != nil {
			kg.logger.Printf("SKIP lat inválido: neighborhood_id=%d, lat=%q, err=%v",
				report.NeighborhoodID, report.Neighborhood.Latitude, err)
			skipped++
			continue
		}

		lon, err := strconv.ParseFloat(strings.ReplaceAll(report.Neighborhood.Longitude, ",", "."), 64)
		if err != nil {
			kg.logger.Printf("SKIP lon inválido: neighborhood_id=%d, lon=%q, err=%v",
				report.NeighborhoodID, report.Neighborhood.Longitude, err)
			skipped++
			continue
		}

		if lat < -23.1 || lat > -22.7 || lon < -47.3 || lon > -46.8 {
			kg.logger.Printf("SKIP fora da bounding box: neighborhood_id=%d, lat=%f, lon=%f",
				report.NeighborhoodID, lat, lon)
			skipped++
			continue
		}

		reportTime, err := time.Parse("2006-01-02", report.ReportDateFormated)
				if t, e := time.Parse(time.RFC3339, report.ReportDateFormated); e == nil {
			reportTime = t
		} else if t, e := time.Parse("2006-01-02 15:04:05", report.ReportDateFormated); e == nil {
			// 2) Tenta "YYYY-MM-DD HH:MM:SS"
			reportTime = t
		} else if t, e := time.Parse("2006-01-02", report.ReportDateFormated); e == nil {
			// 3) Tenta só data "YYYY-MM-DD"
			reportTime = t
		} else {
			kg.logger.Printf("SKIP data inválida: report_id=%d, report_date=%q, err=%v",
				report.ReportID, report.ReportDateFormated, e)
			skipped++
			continue
		}

    	// ... resto igual ...


		category := kg.mapCrimeCategory(report.Crime.CrimeName)
		severity := report.Crime.CrimeWeight
		confidence := kg.calculateConfidence(report)
		incidentID := fmt.Sprintf("rpt_%d", report.ReportID)

		// se já atingimos o máximo de linhas por INSERT, dispara e começa outro
		if len(valueStrings) >= maxRowsPerInsert {
			flushBatch()
		}

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

		paramIndex += paramsPerRow
	}

	// flush final
	flushBatch()

	return processed, skipped
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
// FASE 2.5: MAPEAMENTO CÉLULA → BAIRRO
// ============================================================================

func (kg *KnowledgeBaseGenerator) mapCellsToNeighborhoods(ctx context.Context, db *sql.DB) error {
	query := `
        IF OBJECT_ID('cell_neighborhoods', 'U') IS NULL
        BEGIN
            CREATE TABLE cell_neighborhoods (
                cell_id      VARCHAR(50) PRIMARY KEY,
                neighborhood VARCHAR(100) NOT NULL,
                distance     FLOAT NULL
            );
        END;

        TRUNCATE TABLE cell_neighborhoods;

        INSERT INTO cell_neighborhoods (cell_id, neighborhood, distance)
        SELECT
            c.cell_id,
            nearest.name AS neighborhood,
            nearest.distance
        FROM curated_cells c
        CROSS APPLY (
            SELECT TOP 1
                n.name,
                SQRT(
                    POWER(n.latitude  - c.center_lat, 2) +
                    POWER(n.longitude - c.center_lng, 2)
                ) AS distance
            FROM neighborhoods n
            WHERE n.latitude IS NOT NULL 
              AND n.longitude IS NOT NULL
            ORDER BY 
                POWER(n.latitude  - c.center_lat, 2) +
                POWER(n.longitude - c.center_lng, 2)
        ) AS nearest;
    `
	_, err := db.ExecContext(ctx, query)
	return err
}

// ============================================================================
// FASE 3: ATRIBUIR CÉLULAS AOS INCIDENTES
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
// FASE 3.5: FEATURES MENSAIS
// ============================================================================

func (kg *KnowledgeBaseGenerator) ensureMonthlyFeaturesTable(ctx context.Context, db *sql.DB) error {
	query := `
    IF OBJECT_ID('features_cell_monthly', 'U') IS NULL
    BEGIN
        CREATE TABLE features_cell_monthly (
            cell_id         VARCHAR(50) NOT NULL,
            [year]          INT         NOT NULL,
            [month]         INT         NOT NULL,
            y_count_month   INT         NOT NULL DEFAULT 0,
            lag_1m          INT         NOT NULL DEFAULT 0,
            lag_3m          INT         NOT NULL DEFAULT 0,
            PRIMARY KEY (cell_id, [year], [month])
        );

        CREATE INDEX idx_features_cell_monthly_year_month 
            ON features_cell_monthly([year], [month]);
    END
    `
	_, err := db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela features_cell_monthly: %w", err)
	}
	return nil
}

func (kg *KnowledgeBaseGenerator) generateMonthlyFeatures(ctx context.Context, db *sql.DB) error {
	kg.logger.Println("📅 Gerando features mensais por célula...")

	// Garantir que a tabela existe
	if err := kg.ensureMonthlyFeaturesTable(ctx, db); err != nil {
		return err
	}

	query := `
    ;WITH Months AS (
        SELECT 
            DATEFROMPARTS(YEAR(@start), MONTH(@start), 1) AS month_start,
            DATEFROMPARTS(YEAR(@end), MONTH(@end), 1)     AS month_end
    ),
    MonthSeries AS (
        SELECT month_start AS month_date
        FROM Months
        UNION ALL
        SELECT DATEADD(month, 1, month_date)
        FROM MonthSeries
        CROSS JOIN Months
        WHERE DATEADD(month, 1, month_date) <= (SELECT month_end FROM Months)
    ),
    Cells AS (
        SELECT cell_id
        FROM curated_cells
        WHERE cell_resolution = @cellRes
    ),
    CellMonths AS (
        SELECT 
            c.cell_id,
            YEAR(ms.month_date)  AS [year],
            MONTH(ms.month_date) AS [month]
        FROM Cells c
        CROSS JOIN MonthSeries ms
    ),
    IncidentsByMonth AS (
        SELECT
            ci.cell_id,
            YEAR(ci.occurred_at)  AS [year],
            MONTH(ci.occurred_at) AS [month],
            COUNT(*) AS y_count_month
        FROM curated_incidents ci
        WHERE ci.cell_id IS NOT NULL
        GROUP BY ci.cell_id, YEAR(ci.occurred_at), MONTH(ci.occurred_at)
    ),
    Aggregated AS (
        SELECT
            cm.cell_id,
            cm.[year],
            cm.[month],
            ISNULL(ibm.y_count_month, 0) AS y_count_month
        FROM CellMonths cm
        LEFT JOIN IncidentsByMonth ibm
            ON ibm.cell_id = cm.cell_id
           AND ibm.[year]  = cm.[year]
           AND ibm.[month] = cm.[month]
    ),
    WithLags AS (
        SELECT
            a1.cell_id,
            a1.[year],
            a1.[month],
            a1.y_count_month,

            -- lag_1m: mês anterior
            ISNULL((
                SELECT TOP 1 a0.y_count_month
                FROM Aggregated a0
                WHERE a0.cell_id = a1.cell_id
                  AND DATEADD(month, 1, DATEFROMPARTS(a0.[year], a0.[month], 1)) 
                    = DATEFROMPARTS(a1.[year], a1.[month], 1)
            ), 0) AS lag_1m,

            -- lag_3m: soma dos últimos 3 meses (não incluindo o mês atual)
            ISNULL((
                SELECT SUM(a0.y_count_month)
                FROM Aggregated a0
                WHERE a0.cell_id = a1.cell_id
                  AND DATEFROMPARTS(a0.[year], a0.[month], 1) 
                    >= DATEADD(month, -3, DATEFROMPARTS(a1.[year], a1.[month], 1))
                  AND DATEFROMPARTS(a0.[year], a0.[month], 1) 
                    < DATEFROMPARTS(a1.[year], a1.[month], 1)
            ), 0) AS lag_3m
        FROM Aggregated a1
    )
    MERGE features_cell_monthly AS target
    USING WithLags AS source
        ON target.cell_id = source.cell_id
       AND target.[year]  = source.[year]
       AND target.[month] = source.[month]
    WHEN MATCHED THEN
        UPDATE SET
            y_count_month = source.y_count_month,
            lag_1m        = source.lag_1m,
            lag_3m        = source.lag_3m
    WHEN NOT MATCHED BY TARGET THEN
        INSERT (cell_id, [year], [month], y_count_month, lag_1m, lag_3m)
        VALUES (source.cell_id, source.[year], source.[month],
                source.y_count_month, source.lag_1m, source.lag_3m)
    OPTION (MAXRECURSION 0);
    `

	_, err := db.ExecContext(ctx, query,
		sql.Named("start", kg.config.StartDate),
		sql.Named("end", kg.config.EndDate),
		sql.Named("cellRes", kg.config.CellResolution),
	)
	if err != nil {
		return fmt.Errorf("erro ao gerar features mensais: %w", err)
	}

	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM features_cell_monthly`).Scan(&count); err == nil {
		kg.logger.Printf("✅ Features mensais geradas: %d registros", count)
	} else {
		kg.logger.Println("✅ Features mensais geradas com sucesso")
	}

	return nil
}

// ============================================================================
// FUNÇÕES AUXILIARES (CATEGORIA, CONFIANÇA)
// ============================================================================

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
// FASE 4: FEATURES TEMPORAIS (HORÁRIAS)
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

	if err != nil {
		// 2627 = chave duplicada
		if strings.Contains(err.Error(), "2627") {
			kg.logger.Printf("ℹ️  Registro de qualidade já existe para hoje, atualizando em vez de inserir...")

			updateQuery := `
				UPDATE analytics_quality_reports 
				SET metrics = @p1
				WHERE report_date = CAST(GETDATE() AS DATE)
			`
			if _, errUpdate := db.ExecContext(ctx, updateQuery, string(metricsJSON)); errUpdate != nil {
				kg.logger.Printf("⚠️  Erro ao atualizar analytics_quality_reports: %v", errUpdate)
				return errUpdate
			}
			// sucesso na atualização → não é erro crítico
			return nil
		}

		// outros erros (não de chave duplicada) devem ser propagados
		return err
	}

	kg.logger.Printf("📊 Métricas de qualidade:")
	kg.logger.Printf("   • Total de incidentes: %d", incidentsCount)
	kg.logger.Printf("   • Total de células: %d", cellsCount)
	kg.logger.Printf("   • Total de features: %d", featuresCount)

	return nil
}