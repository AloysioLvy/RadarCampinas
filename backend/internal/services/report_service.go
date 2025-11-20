package services

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReportService defines business operations
// related to crime reports.
type ReportService interface {
	// CreateReport receives a Report object and inserts
	// it into the database, returning an error in case of failure.
	CreateReport(ctx context.Context, r *models.Report) error
	
	// CreateReportsBatch inserts multiple reports in a single transaction
	CreateReportsBatch(ctx context.Context, reports []*models.Report) error
	
	// ProcessReportText processes a single report request
	ProcessReportText(ctx context.Context, req *models.ReportRequest) (*models.Report, error)
	
	// ProcessReportsBatch processes multiple report requests efficiently
	ProcessReportsBatch(ctx context.Context, requests []*models.ReportRequest) ([]*models.Report, error)
	
	// FindOrCreateNeighborhood finds or creates a neighborhood
	FindOrCreateNeighborhood(ctx context.Context, n *models.Neighborhood) (uint, error)
	
	// FindOrCreateCrime finds or creates a crime type
	FindOrCreateCrime(ctx context.Context, c *models.Crime) (uint, error)
}

// reportService is the concrete implementation of ReportService.
// It has the GORM instance to persist data in the database.
type reportService struct {
	db *gorm.DB
	
	// Optional: in-memory cache for frequently accessed data
	neighborhoodCache sync.Map // map[string]uint (key: "lat,lon")
	crimeCache        sync.Map // map[string]uint (key: crime_name)
}

// NewReportService injects the *gorm.DB dependency and returns
// a ReportService instance ready for use.
func NewReportService(db *gorm.DB) ReportService {
	return &reportService{
		db: db,
	}
}

// ============================================================================
// SINGLE OPERATIONS
// ============================================================================

// CreateReport inserts a new report record in the database.
// - Receives the request context for timeout control.
// - r is the pointer to models.Report containing the data to save.
// - Returns an error if the Create operation fails.
func (s *reportService) CreateReport(ctx context.Context, r *models.Report) error {
	if r == nil {
		return errors.New("report cannot be nil")
	}
	
	// Use Clauses for better SQL Server compatibility
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(r).Error
}

// FindOrCreateNeighborhood finds or creates a neighborhood by coordinates.
// Uses GORM's FirstOrCreate for atomic operation.
func (s *reportService) FindOrCreateNeighborhood(ctx context.Context, n *models.Neighborhood) (uint, error) {
	var existing models.Neighborhood

	// Sempre usar WithContext
	result := s.db.WithContext(ctx).
		Where("latitude = ? AND longitude = ?", n.Latitude, n.Longitude).
		First(&existing)

	if result.Error == nil {
		// Copia o registro encontrado para o ponteiro n
		*n = existing
		return existing.NeighborhoodID, nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, result.Error
	}

	// Não existe, criar
	if err := s.db.WithContext(ctx).Create(n).Error; err != nil {
		return 0, err
	}

	return n.NeighborhoodID, nil
}

// FindOrCreateCrime finds or creates a crime type by name.
// Uses GORM's FirstOrCreate for atomic operation.
func (s *reportService) FindOrCreateCrime(ctx context.Context, c *models.Crime) (uint, error) {
	var existing models.Crime

	result := s.db.WithContext(ctx).
		Where("crime_name = ?", c.CrimeName).
		First(&existing)

	if result.Error == nil {
		*c = existing
		return existing.CrimeID, nil
	}

	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return 0, result.Error
	}

	// Não existe, criar
	if err := s.db.WithContext(ctx).Create(c).Error; err != nil {
		return 0, err
	}

	return c.CrimeID, nil
}

// ProcessReportText processes a single report request with transaction support.
 func (s *reportService) ProcessReportText(ctx context.Context, req *models.ReportRequest) (*models.Report, error) {
	// 1) Neighborhood
	neighborhood := &models.Neighborhood{
		Name:               req.Name,
		Latitude:           req.Latitude,
		Longitude:          req.Longitude,
		NeighborhoodWeight: 1,
	}

	neighborhoodID, err := s.FindOrCreateNeighborhood(ctx, neighborhood)
	if err != nil {
		return nil, fmt.Errorf("erro em FindOrCreateNeighborhood: %w", err)
	}
	fmt.Println("NeighborhoodID:", neighborhoodID, "Struct:", neighborhood)
	// 2) Crime
	crime := &models.Crime{
		CrimeName:   req.CrimeName,
		CrimeWeight: req.CrimeWeight,
	}

	crimeID, err := s.FindOrCreateCrime(ctx, crime)
	if err != nil {
		return nil, fmt.Errorf("erro em FindOrCreateCrime: %w", err)
	}
	fmt.Println("CrimeID:", crimeID, "Struct:", crime)
	// 3) Report
	report := &models.Report{
		NeighborhoodID: neighborhoodID,
		CrimeID:        crimeID,
		ReportDate:     req.ReportDate,
	}

	if err := s.CreateReport(ctx, report); err != nil {
		return nil, fmt.Errorf("erro em CreateReport: %w", err)
	}
	fmt.Println("Creating report with NID", neighborhoodID, "CID", crimeID)
	// 4) Preencher structs para resposta (opcional, mas bom)
	report.Neighborhood = *neighborhood
	report.Crime = *crime

	return report, nil
}
// ============================================================================
// BATCH OPERATIONS (PERFORMANCE OPTIMIZATION)
// ============================================================================

// CreateReportsBatch inserts multiple reports in a single batch operation.
// This is much faster than inserting one by one.
func (s *reportService) CreateReportsBatch(ctx context.Context, reports []*models.Report) error {
	if len(reports) == 0 {
		return nil
	}
	
	// GORM's CreateInBatches is optimized for bulk inserts
	// Batch size of 500 is a good balance between memory and performance
	batchSize := 500
	
	return s.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(reports, batchSize).Error
}

// ProcessReportsBatch processes multiple report requests efficiently.
// This method optimizes by:
// 1. Pre-loading all unique neighborhoods and crimes
// 2. Creating them in batch
// 3. Creating all reports in batch
func (s *reportService) ProcessReportsBatch(ctx context.Context, requests []*models.ReportRequest) ([]*models.Report, error) {
	if len(requests) == 0 {
		return nil, nil
	}
	
	var reports []*models.Report
	
	// Use transaction for atomicity
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Collect unique neighborhoods and crimes
		neighborhoodMap := make(map[string]*models.Neighborhood)
		crimeMap := make(map[string]*models.Crime)
		
		for _, req := range requests {
			// Unique key for neighborhood: lat,lon
			nKey := fmt.Sprintf("%s,%s", req.Latitude, req.Longitude)
			if _, exists := neighborhoodMap[nKey]; !exists {
				neighborhoodMap[nKey] = &models.Neighborhood{
					Name:               req.Name,
					Latitude:           req.Latitude,
					Longitude:          req.Longitude,
					NeighborhoodWeight: 1,
				}
			}
			
			// Unique key for crime: crime_name
			if _, exists := crimeMap[req.CrimeName]; !exists {
				crimeMap[req.CrimeName] = &models.Crime{
					CrimeName:   req.CrimeName,
					CrimeWeight: req.CrimeWeight,
				}
			}
		}
		
		// Step 2: Batch create/find neighborhoods
		neighborhoodIDs := make(map[string]uint)
		for key, n := range neighborhoodMap {
			// Use FirstOrCreate for each unique neighborhood
			result := tx.Where("latitude = ? AND longitude = ?", n.Latitude, n.Longitude).
				Attrs(*n).
				FirstOrCreate(n)
			
			if result.Error != nil {
				return fmt.Errorf("failed to process neighborhood %s: %w", key, result.Error)
			}
			neighborhoodIDs[key] = n.NeighborhoodID
		}
		
		// Step 3: Batch create/find crimes
		crimeIDs := make(map[string]uint)
		for name, c := range crimeMap {
			result := tx.Where("crime_name = ?", c.CrimeName).
				Attrs(*c).
				FirstOrCreate(c)
			
			if result.Error != nil {
				return fmt.Errorf("failed to process crime %s: %w", name, result.Error)
			}
			crimeIDs[name] = c.CrimeID
		}
		
		// Step 4: Create all reports
		reports = make([]*models.Report, 0, len(requests))
		for _, req := range requests {
			nKey := fmt.Sprintf("%s,%s", req.Latitude, req.Longitude)
			
			report := &models.Report{
				NeighborhoodID: neighborhoodIDs[nKey],
				CrimeID:        crimeIDs[req.CrimeName],
				ReportDate:     req.ReportDate,
			}
			reports = append(reports, report)
		}
		
		// Batch insert all reports
		if len(reports) > 0 {
			if err := tx.CreateInBatches(reports, 500).Error; err != nil {
				return fmt.Errorf("failed to create reports batch: %w", err)
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	return reports, nil
}