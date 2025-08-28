package services

import (
	"context"

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
	"gorm.io/gorm"
)

// NeighborhoodService defines business operations related to neighborhoods
type NeighborhoodService interface {
	CreateNeighborhood(ctx context.Context, n *models.Neighborhood) error
	GetNeighborhoodByID(ctx context.Context, id uint) (*models.Neighborhood, error)
	GetAllNeighborhoods(ctx context.Context) ([]models.Neighborhood, error)
}

// neighborhoodService is the concrete implementation of NeighborhoodService
type neighborhoodService struct {
	db *gorm.DB
}

// NewNeighborhoodService creates a new instance of NeighborhoodService
func NewNeighborhoodService(db *gorm.DB) NeighborhoodService {
	return &neighborhoodService{db: db}
}

// CreateNeighborhood creates a new neighborhood in the database
func (s *neighborhoodService) CreateNeighborhood(ctx context.Context, n *models.Neighborhood) error {
	return s.db.WithContext(ctx).Create(n).Error
}

// GetNeighborhoodByID retrieves a neighborhood by its ID
func (s *neighborhoodService) GetNeighborhoodByID(ctx context.Context, id uint) (*models.Neighborhood, error) {
	var neighborhood models.Neighborhood
	err := s.db.WithContext(ctx).First(&neighborhood, id).Error
	if err != nil {
		return nil, err
	}
	return &neighborhood, nil
}

// GetAllNeighborhoods retrieves all neighborhoods from the database
func (s *neighborhoodService) GetAllNeighborhoods(ctx context.Context) ([]models.Neighborhood, error) {
	var neighborhoods []models.Neighborhood
	err := s.db.WithContext(ctx).Find(&neighborhoods).Error
	return neighborhoods, err
}