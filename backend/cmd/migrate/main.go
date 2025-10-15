package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Conectar ao banco
	connStr := "host=localhost port=5432 user=seu_usuario password=sua_senha dbname=seu_banco sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Erro ao conectar ao banco:", err)
	}
	defer db.Close()

	// Testar conexão
	if err := db.Ping(); err != nil {
		log.Fatal("Erro ao pingar o banco:", err)
	}

	fmt.Println("✅ Conectado ao banco com sucesso!")

	// Ler o arquivo SQL
	sqlFile := "internal/database/migrations/knowledge_base_schema.sql"
	sqlBytes, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatal("Erro ao ler arquivo SQL:", err)
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
	defer rows.Close()

	fmt.Println("\n📊 Schemas criados:")
	for rows.Next() {
		var schema string
		rows.Scan(&schema)
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
	defer rows2.Close()

	fmt.Println("\n📋 Tabelas criadas:")
	currentSchema := ""
	for rows2.Next() {
		var schema, table string
		rows2.Scan(&schema, &table)
		if schema != currentSchema {
			fmt.Printf("\n  %s:\n", schema)
			currentSchema = schema
		}
		fmt.Printf("    ✓ %s\n", table)
	}

	fmt.Println("\n🎉 Tudo pronto!")
}
