package services

import (
	"context"
	"errors"

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
	"gorm.io/gorm"
)

// ReportService defines business operations
// related to crime reports.
type ReportService interface {
	// CreateReport receives a Report object and inserts
	// it into the database, returning an error in case of failure.
	CreateReport(ctx context.Context, r *models.Report) error
	ProcessReportText(ctx context.Context, req *models.ReportRequest) (*models.Report, error)
	FindOrCreateNeighborhood(ctx context.Context, n *models.Neighborhood) (uint, error)
	FindOrCreateCrime(ctx context.Context, c *models.Crime) (uint, error)
}

// reportService is the concrete implementation of ReportService.
// It has the GORM instance to persist data in the database.
type reportService struct {
	db *gorm.DB
}

// NewReportService injects the *gorm.DB dependency and returns
// a ReportService instance ready for use.
func NewReportService(db *gorm.DB) ReportService {
	return &reportService{db: db}
}

// CreateReport inserts a new report record in the database.
// - Receives the request context for timeout control.
// - r is the pointer to models.Report containing the data to save.
// - Returns an error if the Create operation fails.
func (s *reportService) CreateReport(ctx context.Context, r *models.Report) error {
	// db.WithContext(ctx) associates the context with GORM
	// and Create(r) performs the INSERT in the "Reports" table.
	return s.db.WithContext(ctx).Create(r).Error
}

func (s *reportService) FindOrCreateNeighborhood(ctx context.Context, n *models.Neighborhood) (uint, error) {
	var existingNeighborhood models.Neighborhood
	result := s.db.Where("latitude = ? AND longitude = ?", n.Latitude, n.Longitude).First(&existingNeighborhood)

	if result.Error == nil {
		// Bairro encontrado → SOMAR o peso
		existingNeighborhood.NeighborhoodWeight += n.NeighborhoodWeight

		// Atualizar no banco
		if err := s.db.WithContext(ctx).
			Model(&existingNeighborhood).
			Update("neighborhood_weight", existingNeighborhood.NeighborhoodWeight).Error; err != nil {

			return 0, err
		}

		return existingNeighborhood.NeighborhoodID, nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, result.Error
	}

	// Bairro não encontrado → criar um novo
	if err := s.db.WithContext(ctx).Create(n).Error; err != nil {
		return 0, err
	}

	return n.NeighborhoodID, nil
}

func (s *reportService) FindOrCreateCrime(ctx context.Context, c *models.Crime) (uint, error) {
	// Check if a crime with this name already exists
	var existingCrime models.Crime
	result := s.db.Where("crime_name = ?", c.CrimeName).First(&existingCrime)

	if result.Error == nil {
		// Crime found, return the ID
		return existingCrime.CrimeID, nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Database error
		return 0, result.Error
	}

	// Crime not found, create a new one
	newCrime := models.Crime{
		CrimeName:   c.CrimeName,
		CrimeWeight: c.CrimeWeight,
	}

	if err := s.db.WithContext(ctx).Create(&newCrime).Error; err != nil {
		return 0, err
	}

	return newCrime.CrimeID, nil
}

func (s *reportService) ProcessReportText(ctx context.Context, req *models.ReportRequest) (*models.Report, error) {
	// Create or find neighborhood
	neighborhood := &models.Neighborhood{
		Name:               req.Name,
		Latitude:           req.Latitude,
		Longitude:          req.Longitude,
		NeighborhoodWeight: 1, // Default weight
	}

	neighborhoodID, err := s.FindOrCreateNeighborhood(ctx, neighborhood)
	if err != nil {
		return nil, err
	}

	// Create or find crime
	crime := &models.Crime{
		CrimeName:   req.CrimeName,
		CrimeWeight: req.CrimeWeight,
	}

	crimeID, err := s.FindOrCreateCrime(ctx, crime)
	if err != nil {
		return nil, err
	}

	// Create the report
	report := &models.Report{
		NeighborhoodID: neighborhoodID,
		CrimeID:        crimeID,
		ReportDate:     req.ReportDate,
	}

	if err := s.CreateReport(ctx, report); err != nil {
		return nil, err
	}

	return report, nil
}