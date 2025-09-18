package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"database/sql"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AloysioLvy/TccRadarCampinas/backend/cmd/server"
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

func (c *KnowledgeBaseController) GenerateKnowledgeBaseHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// conectar ao source DB
	sourceDB, err := sql.Open("postgres", c.SourceDSN)
	if err != nil {
		http.Error(w, "Erro conectando ao source DB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer sourceDB.Close()

	// conectar ao target DB
	targetDB, err := pgxpool.New(ctx, c.TargetDSN)
	if err != nil {
		http.Error(w, "Erro conectando ao target DB: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer targetDB.Close()

	config := &main.KnowledgeBaseConfig{
		SourceDB:       sourceDB,
		TargetDB:       targetDB,
		CellResolution: 500, // ou 1000
		BatchSize:      500,
		StartDate:      time.Now().AddDate(-1, 0, 0), // últimos 12 meses
		EndDate:        time.Now(),
	}

	generator := main.NewKnowledgeBaseGenerator(config)

	if err := generator.GenerateKnowledgeBase(ctx); err != nil {
		log.Printf("Erro ao gerar KB: %v", err)
		http.Error(w, "Falha ao gerar a base de conhecimento: "+err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]string{"status": "Base de conhecimento gerada com sucesso!"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
