package main

import (
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

	// Initialize Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Register routes
	api := e.Group("/api/v1")
	// crimeCtrl.Register(api)
	// neighborhoodCtrl.Register(api)
	reportCtrl.Register(api)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
