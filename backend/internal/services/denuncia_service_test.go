package services

import (
    "context"
	"testing"
	"time"

    "github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
    "gorm.io/gorm"
	"gorm.io/driver/sqlite"
)

// ptrString e ptrInt16 ajudam a criar ponteiros para literais
func ptrString(s string) *string { return &s }
func ptrInt16(i int16) *int16 { return &i }


// setupTestDB abre um SQLite em memoria e migra Crime, Bairro e Denuncia
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.FatalF("não foi possivel abrir DB de teste: %v", err)
	}
	// É preciso migrar todas as tabelas envolvidas (por causa de FKs)
	if err := db.AutoMigrate(&models.Crime{}, &models.Bairro{}, &models.Denuncia{}); err != nil {
		t.Fatalf("falha na migração dos modelos: %v", err)
	}
	return db
}

// TestCreateDenuncia_Succes verifica que CreateDenuncia insere o registro
// no banco e mantem os campos corretos
func TestCreateDenuncia_Succes(t *testing.T) {
	db := setupTestDB(t)

	// 1. Semente de crime e bairro para satisfazer as FKs
	crime := models.Crime{IDCrime: 1, NomeCrime: ptrString("Assalto"), PesoCrime: ptrInt16(7)}
	bairro := models.Bairro{IDBairro: 1, Nome: "Centro", Latitude: "0", Longitude: "0", PesoBairro: 5}

	if err := db.Create(&crime).Error; err != nil {
		t.Fatalf("falha ao inserir crime semente: %v", err)
	}
	if err := db.Create(&bairro).Error; err != nil {
		t.Fatalf("falha ao inserir bairro semente: %v", err)
	}

	svc := NewDenunciaService(db)

	// 2. Cria objeto Denuncia
	now := time.Now().UTC().Truncate(time.Second)
	d := &models.Denuncia{
		IDDenuncia: 1,
		IDCrime: crime.IDCrime,
		IDBairro: bairro.IDBairro,
		Data: now,
	}

	// 3. Chama CreateDenuncia
	if err := svc.CreateDenuncia(context.Background(), d); err != nil {
		t.Fatalf("esperava sem erro ao criar denuncia, obteve: %v", err)
	}

	// 4. Busca no banco para verificar inserção correta
	var saved models.Denuncia
	if err := db.First(&saved, "id_denuncia = ?", d.IDDenuncia).Error; err != nil {
		t.Fatalf("falha ao buscar denuncia inserida: %v", err)
	}

	if saved.IDCrime != d.IDCrime || saved.IDBairro != d.IDBairro {
		t.Errorf("campos IDCrime ou IDBairro não batem: got %+v, want %+v", saved, d)
	}

	if !saved.Data.Equal(d.Data) {
		t.Errorf("campo Data não bate: got %v, want %v", saved.Data, d.Data)
	}
}

