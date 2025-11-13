package controllers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/services"
)

type KnowledgeBaseController struct {
	SourceDSN string
	TargetDSN string
	Logger    *log.Logger
}

func NewKnowledgeBaseController(sourceDSN, targetDSN string) *KnowledgeBaseController {
	return &KnowledgeBaseController{
		SourceDSN: sourceDSN,
		TargetDSN: targetDSN,
		Logger:    log.New(os.Stdout, "[KB-CTRL] ", log.LstdFlags|log.Lmsgprefix),
	}
}

// ============================================================================
// HANDLERS
// ============================================================================

// GenerateKnowledgeBaseHandler é o ponto de entrada que dispara a geração da base
func (c *KnowledgeBaseController) GenerateKnowledgeBaseHandler(ctx echo.Context) error {
	c.Logger.Println("📥 Recebida requisição para gerar base de conhecimento")

	background := context.Background()

	// Parse query parameters (opcional)
	cellResolution := 500 // padrão
	if res := ctx.QueryParam("cell_resolution"); res != "" {
		if parsed, err := strconv.Atoi(res); err == nil && (parsed == 500 || parsed == 1000) {
			cellResolution = parsed
		}
	}

	daysBack := 365 // padrão: 1 ano
	if days := ctx.QueryParam("days_back"); days != "" {
		if parsed, err := strconv.Atoi(days); err == nil && parsed > 0 {
			daysBack = parsed
		}
	}

	c.Logger.Printf("⚙️  Parâmetros: cell_resolution=%dm, days_back=%d", cellResolution, daysBack)

	// Conectar no Source DB (legado)
	c.Logger.Println("🔌 Conectando ao Source Database (legado)...")
	sourceDB, err := sql.Open("postgres", c.SourceDSN)
	if err != nil {
		c.Logger.Printf("❌ Erro ao conectar no Source DB: %v", err)
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "Erro ao conectar no Source DB",
			"details": err.Error(),
		})
	}
	defer func() {
		if err := sourceDB.Close(); err != nil {
			c.Logger.Printf("Erro ao fechar Source DB: %v", err)
		}
	}()

	// Testar conexão
	if err := sourceDB.Ping(); err != nil {
		c.Logger.Printf("❌ Source DB não responde: %v", err)
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "Source DB não está acessível",
			"details": err.Error(),
		})
	}
	c.Logger.Println("✅ Source DB conectado")

	// Conectar no Target DB (KB para IA)
	c.Logger.Println("🔌 Conectando ao Target Database (KB)...")
	targetDB, err := pgxpool.New(background, c.TargetDSN)
	if err != nil {
		c.Logger.Printf("❌ Erro ao conectar no Target DB: %v", err)
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "Erro ao conectar no Target DB",
			"details": err.Error(),
		})
	}
	defer targetDB.Close()

	// Testar conexão
	if err := targetDB.Ping(background); err != nil {
		c.Logger.Printf("❌ Target DB não responde: %v", err)
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"error":   "Target DB não está acessível",
			"details": err.Error(),
		})
	}
	c.Logger.Println("✅ Target DB conectado")

	// Configurar gerador
	config := &services.KnowledgeBaseConfig{
		SourceDB:       sourceDB,
		TargetDB:       targetDB,
		CellResolution: cellResolution,
		BatchSize:      500,
		StartDate:      time.Now().AddDate(0, 0, -daysBack),
		EndDate:        time.Now(),
	}

	generator := services.NewKnowledgeBaseGenerator(config)

	// Executar geração
	c.Logger.Println("🚀 Iniciando geração da base de conhecimento...")
	startTime := time.Now()

	if err := generator.GenerateKnowledgeBase(background); err != nil {
		c.Logger.Printf("❌ Erro ao gerar KB: %v", err)
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"error":        "Erro na geração da base de conhecimento",
			"details":      err.Error(),
			"elapsed_time": time.Since(startTime).String(),
		})
	}

	elapsed := time.Since(startTime)
	c.Logger.Printf("✅ Base de conhecimento gerada com sucesso em %s", elapsed)

	return ctx.JSON(http.StatusOK, echo.Map{
		"status":          "success",
		"message":         "Base de conhecimento gerada com sucesso",
		"elapsed_time":    elapsed.String(),
		"cell_resolution": cellResolution,
		"days_processed":  daysBack,
		"start_date":      config.StartDate.Format("2006-01-02"),
		"end_date":        config.EndDate.Format("2006-01-02"),
	})
}

// HealthCheckHandler verifica a saúde do sistema e conectividade com DBs
func (c *KnowledgeBaseController) HealthCheckHandler(ctx echo.Context) error {
	background := context.Background()
	health := echo.Map{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"checks":    echo.Map{},
	}

	checks := health["checks"].(echo.Map)

	// Check Source DB
	sourceDB, err := sql.Open("postgres", c.SourceDSN)
	if err == nil {
		defer func() {
			if err := sourceDB.Close(); err != nil {
				c.Logger.Printf("Erro ao fechar Source DB: %v", err)
			}
		}()
		if err := sourceDB.Ping(); err == nil {
			checks["source_db"] = echo.Map{
				"status":  "ok",
				"message": "Source database is accessible",
			}
		} else {
			checks["source_db"] = echo.Map{
				"status":  "error",
				"message": err.Error(),
			}
			health["status"] = "degraded"
		}
	} else {
		checks["source_db"] = echo.Map{
			"status":  "error",
			"message": err.Error(),
		}
		health["status"] = "degraded"
	}

	// Check Target DB
	targetDB, err := pgxpool.New(background, c.TargetDSN)
	if err == nil {
		defer targetDB.Close()
		if err := targetDB.Ping(background); err == nil {
			// Verificar se schemas existem
			var schemaCount int
			query := `SELECT COUNT(*) FROM pg_namespace WHERE nspname IN ('curated', 'external', 'features', 'analytics')`
			err := targetDB.QueryRow(background, query).Scan(&schemaCount)

			if err == nil && schemaCount == 4 {
				checks["target_db"] = echo.Map{
					"status":  "ok",
					"message": "Target database is accessible and schemas exist",
					"schemas": schemaCount,
				}
			} else if err == nil {
				checks["target_db"] = echo.Map{
					"status":  "warning",
					"message": "Target database accessible but schemas incomplete",
					"schemas": schemaCount,
				}
				health["status"] = "degraded"
			} else {
				checks["target_db"] = echo.Map{
					"status":  "error",
					"message": err.Error(),
				}
				health["status"] = "degraded"
			}
		} else {
			checks["target_db"] = echo.Map{
				"status":  "error",
				"message": err.Error(),
			}
			health["status"] = "degraded"
		}
	} else {
		checks["target_db"] = echo.Map{
			"status":  "error",
			"message": err.Error(),
		}
		health["status"] = "degraded"
	}

	// Determinar status code baseado na saúde
	statusCode := http.StatusOK
	if health["status"] == "degraded" {
		statusCode = http.StatusServiceUnavailable
	}

	return ctx.JSON(statusCode, health)
}

// StatusHandler retorna estatísticas da base de conhecimento
func (c *KnowledgeBaseController) StatusHandler(ctx echo.Context) error {
	background := context.Background()

	targetDB, err := pgxpool.New(background, c.TargetDSN)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{
			"error": "Não foi possível conectar ao Target DB",
		})
	}
	defer targetDB.Close()

	stats := echo.Map{
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Total de incidentes
	var incidentCount int
	err = targetDB.QueryRow(background, "SELECT COUNT(*) FROM curated.incidents").Scan(&incidentCount)
	if err != nil {
		stats["incidents"] = echo.Map{"error": err.Error()}
	} else {
		stats["incidents"] = echo.Map{"count": incidentCount}
	}

	// Total de células
	var cellCount int
	err = targetDB.QueryRow(background, "SELECT COUNT(*) FROM curated.cells").Scan(&cellCount)
	if err != nil {
		stats["cells"] = echo.Map{"error": err.Error()}
	} else {
		stats["cells"] = echo.Map{"count": cellCount}
	}

	// Total de features
	var featureCount int
	err = targetDB.QueryRow(background, "SELECT COUNT(*) FROM features.cell_hourly").Scan(&featureCount)
	if err != nil {
		stats["features"] = echo.Map{"error": err.Error()}
	} else {
		stats["features"] = echo.Map{"count": featureCount}
	}

	// Última execução do pipeline
	var lastExecution time.Time
	var lastStatus string
	query := `
		SELECT started_at, status 
		FROM analytics.pipeline_logs 
		ORDER BY started_at DESC 
		LIMIT 1
	`
	err = targetDB.QueryRow(background, query).Scan(&lastExecution, &lastStatus)
	if err != nil {
		stats["last_execution"] = echo.Map{"error": "Nenhuma execução registrada"}
	} else {
		stats["last_execution"] = echo.Map{
			"timestamp": lastExecution.Format(time.RFC3339),
			"status":    lastStatus,
		}
	}

	// Última métrica de qualidade
	var lastReport string
	query = `
		SELECT metrics 
		FROM analytics.quality_reports 
		ORDER BY report_date DESC 
		LIMIT 1
	`
	err = targetDB.QueryRow(background, query).Scan(&lastReport)
	if err != nil {
		stats["quality_metrics"] = echo.Map{"error": "Nenhum relatório disponível"}
	} else {
		stats["quality_metrics"] = lastReport
	}

	return ctx.JSON(http.StatusOK, stats)
}

// ============================================================================
// ROTAS
// ============================================================================

// Register registra as rotas no Echo API group
func (c *KnowledgeBaseController) Register(g *echo.Group) {
	// Rota principal: gerar base de conhecimento
	g.POST("/knowledge-base/generate", c.GenerateKnowledgeBaseHandler)

	// Health check: verificar saúde do sistema
	g.GET("/knowledge-base/health", c.HealthCheckHandler)

	// Status: estatísticas da KB
	g.GET("/knowledge-base/status", c.StatusHandler)
}
