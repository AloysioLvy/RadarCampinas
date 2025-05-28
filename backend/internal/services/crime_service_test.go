package services

// import (
//     "context"
//     "testing"

//     "github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
//     "gorm.io/driver/sqlite"
//     "gorm.io/gorm"
// )

// // ptrString e ptrInt16 ajudam a criar ponteiros para literais
// func ptrString(s string) *string { return &s }
// func ptrInt16(i int16) *int16 { return &i}

// // setupTestDB abre um SQLite em memória e migra apenas o modelo Crime
// func setupTestDB(t *testing.T) *gorm.DB {
// 	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
// 	if err != nil {
// 		t.Fatalf("não foi possivel abrir DB de teste: %v", err)
// 	}
// 	if err := db.AutoMigrate(&models.Crime{}); err != nil {
// 		t.Fatalf("falha na migração do modelo Crime: %v", err)
// 	}
// 	return db
// }

// // TestListCrimes_Empty verifica que ListCrimes retorna slice vazio
// // quando não há registros na tabela
// func TestListCrimes_Emp(t *testing.T) {
// 	db := setupTestDB(t)
// 	svc := NewCrimeService(db)

// 	crimes, err := svc.ListCrimes(context.Background())
// 	if err != nil {
// 		t.Fatalf("esperava sem erro, obteve: %v", err)
// 	}
// 	if len(crimes) != 0 {
// 		t.Errorf("esperava 0 crimes, obteve: %d", len(crimes))
// 	}
// }

// // TestListCrimes_WithData verifica que ListCrimes retorna todos os registros
// // previamente inseridos na tabela.
// func TestListCrimes_WithData(t *testing.T) {
// 	db := setupTestDB(t)
// 	// Semente de dados
// 	c1 := models.Crime{IDCrime: 1, NomeCrime: ptrString("Furto"), PesoCrime: ptrInt16(5)}
// 	c2 := models.Crime{IDCrime: 2, NomeCrime: ptrString("Homicidio"), PesoCrime: ptrInt16(10)}
// 	if err := db.Create(&c1).Error; err != nil {
// 		t.Fatalf("falha ao inserir crime1: %v", err)
// 	}

// 	svc := NewCrimeService(db)
// 	crimes, err := svc.ListCrimes(context.Background())
// 	if err != nil {
// 		t.Fatalf("esperava sem erro, obteve: %v", err)
// 	}
// 	if len(crimes) != 2 {
// 		t.Errorf("esperava 2 crimes, obteve: %d", len(crimes))
// 	}

// 	// Verifica se IDs retornados batem com os inseridos
// 	found := map[int64]bool{}
// 	for _, c := range crimes {
// 		found[c.IDCrime] = true
// 	}
// 	if !found[1] || !found[2] {
// 		t.Errorf("esperava encontrar crimes com IDs 1 e 2, obteve: %v", crimes)
// 	}
// }