package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string
	DBTimezone string
}

func Load() (*Config, error) {
	// carrega .env em dev

	_ = godotenv.Load("/Users/lourenco.diogo/Documents/GitHub/RadarCampinas/.env.local")
	_ = godotenv.Load("/Users/soothsayer/Documents/GitHub/TccRadarCampinas/.env.local")
	_ = godotenv.Load("/Users/u24479/Desktop/TccRadarCampinas/TccRadarCampinas/.env.local")

	
	cfg := &Config{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),
		DBTimezone: os.Getenv("DB_TIMEZONE"),
	}

	fmt.Printf("Config carregada: %+v\n", cfg)

	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBUser == "" || cfg.DBName == "" {
		return nil, fmt.Errorf("Variaveis de ambiente de DB não configuradas")
	}
	return cfg, nil
}
