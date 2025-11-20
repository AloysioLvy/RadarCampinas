package main

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/config"
	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/controllers"
	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/database"
	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database (já está usando SQL Server via GORM)
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate database schemas
	if err := db.AutoMigrate(
		&models.Report{},
		&models.Crime{},
		&models.Neighborhood{},

		// Tabelas da Knowledge Base - Curated
		&models.CuratedIncident{},
		&models.CuratedCell{},

		// Tabelas da Knowledge Base - External
		&models.ExternalHoliday{},

		// Tabelas da Knowledge Base - Features
		&models.FeaturesCellHourly{},

		// Tabelas da Knowledge Base - Analytics
		&models.AnalyticsQualityReport{},
		&models.AnalyticsPipelineLog{},

		// Tabela de migrations
		&models.SchemaMigration{},

		// Predict Time
		&models.PredictCrime{},
	); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Initialize services
	reportSvc := services.NewReportService(db)

	// Create controllers
	reportCtrl := controllers.NewReportController(reportSvc)

	// DSN para SQL Server no Knowledge Base Controller
	sourceDSN := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
	targetDSN := sourceDSN // Mesmo banco para source e target

	kbController := controllers.NewKnowledgeBaseController(sourceDSN, targetDSN)

	// Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Register routes
	api := e.Group("/api/v1")

	// Registrar rotas do report controller
	reportCtrl.Register(api)

	// Registrar rotas do KB controller
	log.Println("🔧 Registrando rotas do Knowledge Base...")
	api.POST("/knowledge-base/generate", kbController.GenerateKnowledgeBaseHandler)
	api.GET("/knowledge-base/health", kbController.HealthCheckHandler)
	api.GET("/knowledge-base/status", kbController.StatusHandler)

	// Rota de teste
	api.GET("/kb-test", func(c echo.Context) error {
		return c.JSON(200, echo.Map{"message": "KB Test route works!"})
	})

	// Listar todas as rotas registradas
	log.Println("📍 Rotas registradas no Echo:")
	for _, route := range e.Routes() {
		log.Printf("   %s %s", route.Method, route.Path)
	}

	// Start server
	log.Println("🚀 Servidor iniciando na porta :8080")
	e.Logger.Fatal(e.Start(":8080"))
}