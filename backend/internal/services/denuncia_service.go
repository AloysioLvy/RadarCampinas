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
	ProcessarDenunciaTexto(ctx context.Context, req *models.DenunciaRequest) (*models.Denuncia, error)
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

// Processa uma denuncia em texto livre
func (s *denunciaService) ProcessarDenunciaTexto(ctx context.Context, req *models.DenunciaRequest) (*models.Denuncia, error) {
	// Inicia uma transação para garantir a consistencia dos dados
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	// Em caso de erro, faz rollback da transacao
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Cria ou encontra o bairro com base nas coordenadas
	bairro := models.Bairro{
		Nome:      req.Nome,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	}

	// Verifica se ja existe um bairro com essas coordenadas
	result := tx.Where("latitude = ? AND longitude = ?", req.Latitude, req.Longitude).First(&bairro)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		tx.Rollback()
		return nil, result.Error
	}

	// Se nao encontrou, cria um novo bairro
	if result.Error == gorm.ErrRecordNotFound {
		if err := tx.Create(&bairro).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 2 . Cria ou encontra o crime com base no tipo e peso
	crime := models.Crime{
		TipoDeCrime: req.TipoDeCrime,
		PesoCrime:   req.PesoCrime,
	}

	// Verifica se ja existe um crime com esse tipo ??

	// Se nao, cria novo crime
	if result.Error == gorm.ErrRecordNotFound {
		if err := tx.Create(&crime).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// 3.Cria a denuncia relacionando bairro e crime
	denuncia := models.Denuncia{
		BairroID:  bairro.BairroID,
		CrimeID:   crime.CrimeID,
		DataCrime: req.DataCrime,
	}

	if err := tx.Create(&denuncia).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 4. Carrega os relacionamentos para retornar a denuncia completa
	if err := tx.Preload("Bairro").Preload("Crime").First(&denuncia, denuncia.DenunciaID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit da transacao
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &denuncia, nil
}
