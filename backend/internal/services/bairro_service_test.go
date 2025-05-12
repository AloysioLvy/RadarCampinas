package services

import (
    "context"
	"testing"
	"time"

    "github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
    "gorm.io/gorm"
	"gorm.io/driver/sqlite"
)

// setupTestDB abre um SQLite em memoria e migra apenas o modelo Bairro
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("não foi possivel abrir DB de teste: %v", err)
	}
	if err := db.AutoMigrate(&models.Bairro{}); err != nil {
		t.Fatalf("falha na migração do modelo Bairro: %v", err)
	}
	return db
}

// TestListBairros_Empty verifica que ListBairros retorna slice vazio
// quando não há registros na tabela.
func TestListBairros_Empty(t *testing.T) {
	db := setupTestDB(t)
	svc := NewBairroService(db)

	bairros, err := svc.ListBairros(context.Background())
	if err != nil {
		t.Fatalf("esperava sem erro, obteve: %v", err)
	}
	if len(bairros) != 0 {
		t.Errorf("esperava 0 bairros, obteve: %d", len(bairros))
	}
}

// TestListBairros_WithData verifica que ListBairros retorna todos os registros 
// previamente inseridos na tabela.
func TestListBairros_WithData(t *testing.T) {
	db := setupTestDB(t)
	// Semente de dados
	b1 := models.Bairro{
		IDBairro: 1,
		Nome: "Centro",
		Latitude: "-23.551",
		Longitude: "-46.633",
		PesoBairro: 10,
	}
	b2 := models.Bairro {
		IDBairro: 2,
		Nome: "Jardins",
		Latitude: "-23.574",
		Longitude: "-46.661",
		PesoBairro: 20,
	}
	if err := db.Create(&b1).Error; err != nil {
		t.Fatalf("falha ao inserir bairro1: %v", err)
	}
	if err := db.Create(&b2).Error; err != nil {
		t.Fatalf("falha ao inserir bairro2: %v", err)
	}

	svc := NewBairroService(db)
	bairros, err := svc.ListBairros(context.Background())
	if err != nil {
		t.Fatalf("esperava sem erro, obteve: %v", err)
	}
	if len(bairros) != 2 {
		t.Errorf("esperava 2 bairros, obteve: %d", len(bairros))
	}

	// Verifica se IDs retornados batem com os inseridos
	found := map[int64]bool{}
	for _, b := range bairros {
		found[b.IDBairro] = true
	}
	if !found[1] || !found[2] {
		t.Errorf("esperava encontrar bairros com IDs 1 e 2, obteve: %v", bairros)
	}
}