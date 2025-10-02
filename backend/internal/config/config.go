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
	if err := godotenv.Load(".env.local"); err == nil {
		fmt.Println("Carregado .env.local")
	} else {
		_ = godotenv.Load(".env.docker")
	}

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
		return nil, fmt.Errorf("Variaveis de ambiente de DB não configuradas: %+v", cfg)
	}
	return cfg, nil
}
