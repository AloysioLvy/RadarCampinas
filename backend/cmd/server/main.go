package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate database schemas
	if err := db.AutoMigrate(&models.Report{}, &models.Crime{}, &models.Neighborhood{}); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Initialize services
	// crimeSvc := services.NewCrimeService(db)
	reportSvc := services.NewReportService(db)
	// neighborhoodSvc := services.NewNeighborhoodService(db)

	// Create controllers
	//crimeCtrl := controllers.NewCrimeController(crimeSvc)
	reportCtrl := controllers.NewReportController(reportSvc)
	//neighborhoodCtrl := controllers.NewNeighborhoodController(neighborhoodSvc)
	kbController := controllers.NewKnowledgeBaseController(
		"postgres://user:pass@localhost:5432/source_db?sslmode=disable",
		"postgres://user:pass@localhost:5433/radar_campinas?sslmode=disable")
	

	// Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Register routes
	api := e.Group("/api/v1")
	// crimeCtrl.Register(api)
	// neighborhoodCtrl.Register(api)
	reportCtrl.Register(api)
	kbController.Register(api)
	
	// Start server
	e.Logger.Fatal(e.Start(":8080"))

	// Configuração (ajuste conforme seu ambiente)
	sourceDB, err := sql.Open("postgres", "postgres://user:pass@localhost/source_db?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer sourceDB.Close()

	targetDB, err := pgxpool.New(context.Background(), "postgres://user:pass@localhost/radar_campinas?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer targetDB.Close()

	config := &KnowledgeBaseConfig{
		SourceDB:       sourceDB,
		TargetDB:       targetDB,
		CellResolution: 500, // 500 metros
		BatchSize:      1000,
		StartDate:      time.Now().AddDate(-1, 0, 0), // último ano
		EndDate:        time.Now(),
	}

	generator := NewKnowledgeBaseGenerator(config)

	ctx := context.Background()
	if err := generator.GenerateKnowledgeBase(ctx); err != nil {
		log.Fatalf("Erro na geração da base de conhecimento: %v", err)
	}

	log.Println("Base de conhecimento gerada com sucesso!")
}
