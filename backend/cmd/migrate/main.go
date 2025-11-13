package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/AloysioLvy/TccRadarCampinas/backend/internal/database/migrations"
	_ "github.com/lib/pq"
)

func main() {
	// Conectar ao banco
	connStr := "host=localhost port=5432 user=seu_usuario password=sua_senha dbname=seu_banco sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco:", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Erro ao fechar conexão: %v", err)
		}
	}()

	// Testar conexão
	if err := db.Ping(); err != nil {
		log.Fatal("Erro ao pingar o banco:", err)
	}

	fmt.Println("✅ Conectado ao banco com sucesso!")

	// Ler o arquivo SQL do embed
	sqlBytes, err := migrations.Files.ReadFile("knowledge_base_schema.sql")
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

	// Verificar schemas criados
	rows, err := db.Query(`
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name IN ('curated', 'external', 'features', 'analytics')
		ORDER BY schema_name
	`)
	if err != nil {
		log.Fatal("Erro ao verificar schemas:", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Erro ao fechar rows: %v", err)
		}
	}()

	fmt.Println("\n📊 Schemas criados:")
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			log.Printf("Erro ao escanear schema: %v", err)
			continue
		}
		fmt.Printf("  ✓ %s\n", schema)
	}

	// Verificar tabelas criadas
	rows2, err := db.Query(`
		SELECT table_schema, table_name 
		FROM information_schema.tables 
		WHERE table_schema IN ('curated', 'external', 'features', 'analytics')
		ORDER BY table_schema, table_name
	`)
	if err != nil {
		log.Fatal("Erro ao verificar tabelas:", err)
	}
	defer func() {
		if err := rows2.Close(); err != nil {
			log.Printf("Erro ao fechar rows2: %v", err)
		}
	}()

	fmt.Println("\n📋 Tabelas criadas:")
	currentSchema := ""
	for rows2.Next() {
		var schema, table string
		if err := rows2.Scan(&schema, &table); err != nil {
			log.Printf("Erro ao escanear tabela: %v", err)
			continue
		}
		if schema != currentSchema {
			fmt.Printf("\n  %s:\n", schema)
			currentSchema = schema
		}
		fmt.Printf("    ✓ %s\n", table)
	}

	fmt.Println("\n🎉 Tudo pronto!")
}
