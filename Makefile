.PHONY: run lint build tidy

# Cores para output
CYAN=\033[0;36m
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m # No Color

# Variáveis
APP_NAME=radar-campinas-kb
GO_FILES=$(shell find . -name '*.go' -not -path './vendor/*')

help: ## Mostra esta mensagem de ajuda
	@echo "$(CYAN)╔═══════════════════════════════════════════════════════════╗$(NC)"
	@echo "$(CYAN)║     🔮 Radar Campinas - Makefile Commands               ║$(NC)"
	@echo "$(CYAN)╚═══════════════════════════════════════════════════════════╝$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(GREEN)%-20s$(NC) %s\n", $$1, $$2}'

build: ## Compilar aplicação Go
	@echo "$(CYAN)🏗️  Compilando aplicação...$(NC)"
	go build ./...
	@echo "$(GREEN)✅ Compilação concluída: bin/$(APP_NAME)$(NC)"

run: ## Executar aplicação
	@echo "$(CYAN)🚀 Iniciando servidor...$(NC)"
	go run ./backend/cmd/server/main.go

test: ## Executar testes
	@echo "$(CYAN)🧪 Executando testes...$(NC)"
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	@echo "$(GREEN)✅ Testes concluídos$(NC)"

test-coverage: ## Executar testes com coverage report
	@echo "$(CYAN)📊 Gerando relatório de cobertura...$(NC)"
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	go tool cover -html=coverage.txt -o coverage.html
	@echo "$(GREEN)✅ Relatório gerado: coverage.html$(NC)"

lint: ## Executar linter
	@echo "$(CYAN)🔍 Executando golangci-lint...$(NC)"
	golangci-lint run --timeout=5m

clean: ## Limpar arquivos gerados
	@echo "$(CYAN)🧹 Limpando arquivos...$(NC)"
	rm -rf bin/
	rm -f coverage.txt coverage.html
	@echo "$(GREEN)✅ Limpeza concluída$(NC)"

docker-up: ## Subir containers Docker
	@echo "$(CYAN)🐳 Subindo containers...$(NC)"
	docker-compose up -d
	@echo "$(GREEN)✅ Containers iniciados$(NC)"
	@echo "$(YELLOW)Source DB: localhost:5433$(NC)"
	@echo "$(YELLOW)Target DB: localhost:5432$(NC)"
	@echo "$(YELLOW)pgAdmin: http://localhost:5050$(NC)"

docker-down: ## Parar containers Docker
	@echo "$(CYAN)🐳 Parando containers...$(NC)"
	docker-compose down
	@echo "$(GREEN)✅ Containers parados$(NC)"

docker-logs: ## Ver logs dos containers
	docker-compose logs -f

migrate: ## Aplicar migrations manualmente
	@echo "$(CYAN)🗄️  Aplicando migrations...$(NC)"
	./scripts/apply_migrations.sh
	@echo "$(GREEN)✅ Migrations aplicadas$(NC)"

kb-generate: ## Gerar base de conhecimento
	@echo "$(CYAN)🔮 Gerando base de conhecimento...$(NC)"
	./scripts/run_kb_generation.sh
	@echo "$(GREEN)✅ Base de conhecimento gerada$(NC)"

kb-generate-fast: ## Gerar KB com menos dados (90 dias, 1000m)
	@echo "$(CYAN)⚡ Gerando base de conhecimento (modo rápido)...$(NC)"
	./scripts/run_kb_generation.sh --days-back=90 --cell-resolution=1000

health: ## Verificar saúde do sistema
	@echo "$(CYAN)🏥 Verificando saúde...$(NC)"
	@curl -s http://localhost:8080/api/v1/knowledge-base/health | python3 -m json.tool

status: ## Ver status da KB
	@echo "$(CYAN)📊 Status da base de conhecimento:$(NC)"
	@curl -s http://localhost:8080/api/v1/knowledge-base/status | python3 -m json.tool

deps: ## Instalar dependências
	@echo "$(CYAN)📦 Instalando dependências...$(NC)"
	go mod download
	go mod tidy
	@echo "$(GREEN)✅ Dependências instaladas$(NC)"

setup: deps docker-up migrate ## Setup completo (deps + docker + migrate)
	@echo "$(GREEN)🎉 Setup completo!$(NC)"
	@echo "$(YELLOW)Execute 'make run' para iniciar o servidor$(NC)"

fmt: ## Formatar código
	@echo "$(CYAN)💅 Formatando código...$(NC)"
	go fmt ./...
	@echo "$(GREEN)✅ Código formatado$(NC)"

vet: ## Executar go vet
	@echo "$(CYAN)🔍 Executando go vet...$(NC)"
	go vet ./...
	@echo "$(GREEN)✅ Verificação concluída$(NC)"

check: fmt vet lint test ## Executar todas as verificações
	@echo "$(GREEN)✅ Todas as verificações passaram!$(NC)"

.DEFAULT_GOAL := help
