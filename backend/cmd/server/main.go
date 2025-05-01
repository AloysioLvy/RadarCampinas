package main

import (
	"log"
	"github.com/gin-gonic/gin" // talvez mudar para echo
	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/config"
	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/database"
	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/routes"
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
	if err: = db.AutoMigrate(&models.Crime{})
}