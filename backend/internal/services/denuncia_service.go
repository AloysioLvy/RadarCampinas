package services

import (
    "context"

    "github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
    "gorm.io/gorm"
)

// DenunciaService define as operações de negócio
// relacionadas a denúncias de crimes.
type DenunciaService interface {
    // CreateDenuncia recebe um objeto Denuncia e insere
    // no banco de dados, retornando erro em caso de falha.
    CreateDenuncia(ctx context.Context, d *models.Denuncia) error
}

// denunciaService é a implementação concreta de DenunciaService.
// Possui a instância de GORM para persitir dados no banco.
type denunciaService struct {
    db *gorm.DB
}

// NewDenunciaService injeta a dependência *gorm.DB e devolve
// uma instância de DenunciaService pronta para uso.
func NewDenunciaService(db *gorm.DB) DenunciaService {
    return &denunciaService{db: db}
}

// CreateDenuncia insere um novo registro de denúncia no banco.
// - Recebe o contexto da requisição para controle de timeout.
// - d é o ponteiro para models.Denuncia contendo os dados a salvar.
// - Retorna um erro se a operação de Create falhar.
func (s *denunciaService) CreateDenuncia(ctx context.Context, d *models.Denuncia) error {
    // db.WithContext(ctx) associa o contexto ao GORM
    // e Create(d) realiza o INSERT na tabela "Denuncia".
    return s.db.WithContext(ctx).Create(d).Error
}