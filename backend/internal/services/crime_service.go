package services

import (
    "context"

    "github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
    "gorm.io/gorm"
)

// CrimeService define o contrato (interface) para operações
// relacionadas à entidade Crime no sistema.
// Aqui declaramos os métodos que qualquer implementação deve oferecer.
type CrimeService interface {
    // ListCrimes retorna todos os registros de crimes
    // e um erro caso algo falhe na recuperação no banco.
    ListCrimes(ctx context.Context) ([]models.Crime, error)
}

// crimeService é a implementação concreta de CrimeService.
// Ela encapsula a dependência do GORM para acessar o banco.
type crimeService struct {
    db *gorm.DB
}

// NewCrimeService é a função fábrica que injeta a dependência
// *gorm.DB e devolve uma instância de CrimeService pronta para uso.
func NewCrimeService(db *gorm.DB) CrimeService {
    return &crimeService{db: db}
}

// ListCrimes recupera todos os crimes do banco de dados.
// - Recebe um contexto (ctx) para permitir controle de timeout,
//   cancellations ou trace.
// - Usa db.WithContext(ctx) para passar o contexto ao GORM.
// - Retorna um slice de models.Crime ou erro.
func (s *crimeService) ListCrimes(ctx context.Context) ([]models.Crime, error) {
    var crimes []models.Crime

    // Executa a consulta sem filtros: SELECT * FROM "Crimes"
    if err := s.db.WithContext(ctx).Find(&crimes).Error; err != nil {
        // Se houver erro no DB, retorna nil slice e o próprio erro.
        return nil, err
    }

    // Retorna o slice preenchido e nil (sem erro)
    return crimes, nil
}