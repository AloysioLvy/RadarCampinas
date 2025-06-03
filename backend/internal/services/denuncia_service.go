package services

import (
	"context"
	"errors"

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
	GetDenuncias(ctx context.Context) ([]models.Denuncia, error)
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

	// Busca ou cria bairro
	var bairro models.Bairro
	err := tx.Where("latitude = ? AND longitude = ?", req.Latitude, req.Longitude).First(&bairro).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 1. Cria ou encontra o bairro com base nas coordenadas
		bairro := models.Bairro{
			Nome:       req.Nome,
			Latitude:   req.Latitude,
			Longitude:  req.Longitude,
			PesoBairro: 0,
		}
		if err := tx.Create(&bairro).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else if err != nil {
		tx.Rollback()
		return nil, err
	}

	var crime models.Crime
	err = tx.Where("tipo_de_crime = ? AND peso_crime = ?", req.TipoDeCrime, req.PesoCrime).First(&crime).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		crime = models.Crime{
			TipoDeCrime: req.TipoDeCrime,
			PesoCrime:   req.PesoCrime,
		}
		if err := tx.Create(&crime).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else if err != nil {
		tx.Rollback()
		return nil, err
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

// GetDenuncias busca todas as denuncias com suas associações
func (s *denunciaService) GetDenuncias(ctx context.Context) ([]models.Denuncia, error) {
	var denuncias []models.Denuncia

	err := s.db.WithContext(ctx).Preload("Bairro").Preload("Crime").Find(&denuncias).Error
	if err != nil {
		return nil, err
	}

	return denuncias, nil
}
