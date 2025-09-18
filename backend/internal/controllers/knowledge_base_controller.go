package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"database/sql"
	"log"

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/services"
)

type KnowledgeBaseController struct {
	SourceDSN string
	TargetDSN string
}

func NewKnowledgeBaseController(sourceDSN, targetDSN string) *KnowledgeBaseController {
	return &KnowledgeBaseController{
		SourceDSN: sourceDSN,
		TargetDSN: targetDSN,
	}
}

// ✅ Ponto de entrada: rota que dispara geração da base
func (c *KnowledgeBaseController) GenerateKnowledgeBaseHandler(ctx echo.Context) error {
	background := context.Background()

	// Conectar no Source DB
	sourceDB, err := sql.Open("postgres", c.SourceDSN)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{"error": "Erro Source DB: " + err.Error()})
	}
	defer sourceDB.Close()

	// Conectar no Target DB
	targetDB, err := pgxpool.New(background, c.TargetDSN)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, echo.Map{"error": "Erro Target DB: " + err.Error()})
	}
	defer targetDB.Close()

	config := &services.KnowledgeBaseConfig{
		SourceDB:       sourceDB,
		TargetDB:       targetDB,
		CellResolution: 500,
		BatchSize:      500,
		StartDate:      time.Now().AddDate(-1, 0, 0),
		EndDate:        time.Now(),
	}
	generator := services.NewKnowledgeBaseGenerator(config)

	if err := generator.GenerateKnowledgeBase(background); err != nil {
		log.Printf("Erro ao gerar KB: %v", err)
		return ctx.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, echo.Map{"status": "Base de conhecimento gerada com sucesso"})
}

// ✅ Método padrão para registrar a rota no Echo API group
func (c *KnowledgeBaseController) Register(g *echo.Group) {
	g.POST("/knowledge-base/generate", c.GenerateKnowledgeBaseHandler)
}
