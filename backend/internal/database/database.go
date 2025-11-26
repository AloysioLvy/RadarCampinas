package database

import (
	"fmt"
	"log"

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/config"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	// Construir DSN para SQL Server
	dsn := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%s?database=%s",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		DisableForeignKeyConstraintWhenMigrating: true, // 👈 ESSENCIAL
	})

	if err != nil {
		return nil, err
	}

	log.Println("✅ Conectado ao SQL Server com sucesso!")
	return db, nil
}