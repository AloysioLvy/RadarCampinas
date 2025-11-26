package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/database/migrations"
	_ "github.com/denisenkom/go-mssqldb"
)

func main() {
	dsn := "sqlserver://BD24452:BD24452@regulus.cotuca.unicamp.br:1433?database=BD24452"
	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco:", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Erro ao fechar conexão: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatal("Erro ao pingar o banco:", err)
	}

	fmt.Println("✅ Conectado ao SQL Server com sucesso!")

	// Ler o arquivo SQL do embed
	sqlBytes, err := migrations.Files.ReadFile("knowledge_base_schema_sqlserver.sql")
	if err != nil {
		log.Fatal("Erro ao ler arquivo SQL embutido:", err)
	}

	fmt.Println("📄 Arquivo SQL lido com sucesso!")
	fmt.Println("🚀 Executando migration...")

	// Executar o SQL
	_, err = db.Exec(string(sqlBytes))
	if err != nil {
		log.Fatal("❌ Erro ao executar migration:", err)
	}

	fmt.Println("✅ Migration executada com sucesso!")

	// Verificar tabelas criadas
	rows, err := db.Query(`
		SELECT TABLE_NAME 
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = 'BD24452'
		AND TABLE_NAME LIKE 'curated_%' 
		   OR TABLE_NAME LIKE 'external_%'
		   OR TABLE_NAME LIKE 'features_%'
		   OR TABLE_NAME LIKE 'analytics_%'
		ORDER BY TABLE_NAME
	`)
	if err != nil {
		log.Fatal("Erro ao verificar tabelas:", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Erro ao fechar rows: %v", err)
		}
	}()

	fmt.Println("\n📋 Tabelas criadas:")
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			log.Printf("Erro ao escanear tabela: %v", err)
			continue
		}
		fmt.Printf("  ✓ %s\n", table)
	}

	fmt.Println("\n🎉 Tudo pronto!")
}