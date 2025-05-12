package services

import (
    "context"

    "github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
    "gorm.io/gorm"
)

// BairroService define o contrato (interface) para operações
// relacionadas à entidade Bairro no sistema.
type BairroService interface {
    // ListBairros retorna todos os registros de bairros
    // e um erro caso algo falhe na recuperação no banco.
    ListBairros(ctx context.Context) ([]models.Bairro, error)
}

// bairroService é a implementação concreta de BairroService.
// Contém a instância de GORM para acessar o banco.
type bairroService struct {
    db *gorm.DB
}

// NewBairroService injeta a dependência *gorm.DB e retorna
// uma instância de BairroService pronta para uso.
func NewBairroService(db *gorm.DB) BairroService {
    return &bairroService{db: db}
}

// ListBairros consulta o banco de dados para obter todos os bairros.
// - O parâmetro ctx permite controlar o fluxo da requisição.
// - Retorna um slice de models.Bairro ou um erro.
func (s *bairroService) ListBairros(ctx context.Context) ([]models.Bairro, error) {
    var bairros []models.Bairro

    // Executa SELECT * FROM "Bairros"
    if err := s.db.WithContext(ctx).Find(&bairros).Error; err != nil {
        // Em caso de falha na query, retorna slice nil e o erro.
        return nil, err
    }

    // Retorna o slice preenchido e sem erro
    return bairros, nil
}