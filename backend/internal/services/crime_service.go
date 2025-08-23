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

type crimeService struct {
    db *gorm.DB
}

func NewCrimeService(db *gorm.DB) CrimeService {
    return &crimeService{db: db}
}


func (s *crimeService) ListCrimes(ctx context.Context) ([]models.Crime, error) {
    var crimes []models.Crime

    // Execute consult with 0 filter: SELECT * FROM "Crimes"
    if err := s.db.WithContext(ctx).Find(&crimes).Error; err != nil {
        return nil, err
    }

    // Return slice and nil 
    return crimes, nil
}