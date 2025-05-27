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
	// Carregar as configs
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Falha ao carregar configs: %v", err)
	}

	// Conectar ao banco
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Falha ao conectar banco de dados: %v", err)
	}

	// Auto-migrate -> a estudar (provisorio)
	if err := db.AutoMigrate(&models.Crime{}, &models.Bairro{}, &models.Denuncia{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	// 4. Instancia serviços
	// crimeSvc := services.NewCrimeService(db)
	denunciaSvc := services.NewDenunciaService(db)
	// bairroSvc := services.NewBairroService(db)

	// 5. Cria controllers
	//crimeCtrl := controllers.NewCrimeController(crimeSvc)
	denunciaCtrl := controllers.NewDenunciaController(denunciaSvc)
	//bairroCtrl := controllers.NewBairroController(bairroSvc)

	// 6. Inicializa Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 7. Registra rotas
	api := e.Group("/api/v1")
	// crimeCtrl.Register(api)
	// bairroCtrl.Register(api)
	denunciaCtrl.Register(api)

	// 8. Roda Servidor
	e.Logger.Fatal(e.Start(":8080"))
}
