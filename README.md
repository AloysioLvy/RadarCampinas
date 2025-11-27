# Tcc Radar Campinas
A segurança urbana é uma preocupação em Campinas, onde a falta de informações expõe moradores a riscos. Nosso app usa IA para analisar dados criminais e gerar um mapa dinâmico de áreas de risco, ajudando na tomada de decisões seguras. Com alto valor social, a solução busca proteger a população ao fornecer insights sobre a criminalidade local.

-Tela inicial
![RadarCampinasmapa](https://github.com/user-attachments/assets/5b762e73-a7b6-46fb-a44f-f96df4fbe90a)
-Chat denúncia 
![image](https://github.com/user-attachments/assets/9cdaa70b-8730-445a-9aef-791f059af30e)


# 🔮 Radar Campinas - Knowledge Base Generator

<div align="center">

![Status](https://img.shields.io/badge/status-active-success.svg)
![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)
![PostgreSQL](https://img.shields.io/badge/postgresql-15+-blue.svg)
![PostGIS](https://img.shields.io/badge/PostGIS-3.4+-green.svg)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

**Sistema de geração de base de conhecimento para análise preditiva de criminalidade em Campinas/SP**

[Arquitetura](docs/ARCHITECTURE.md) • [GitHub Actions](.github/workflows/knowledge-base-pipeline.yml) • [Migrations](internal/database/migrations/)

</div>

---

## 📋 Índice

- [Visão Geral](#-visão-geral)
- [Características](#-características)
- [Pré-requisitos](#-pré-requisitos)
- [Configuração Rápida](#-configuração-rápida)
- [Configuração dos Dois Bancos](#️-configuração-dos-dois-bancos)
- [Como Executar](#-como-executar)
- [Testando a Rota](#-testando-a-rota)
- [Visualizando o Pipeline](#-visualizando-o-pipeline-no-github)
- [Estrutura do Projeto](#-estrutura-do-projeto)
- [API Endpoints](#-api-endpoints)
- [Troubleshooting](#-troubleshooting)
- [Contribuindo](#-contribuindo)

---

## 🎯 Visão Geral

O **Radar Campinas Knowledge Base Generator** é um sistema ETL (Extract, Transform, Load) que processa dados históricos de criminalidade e os transforma em uma base de conhecimento otimizada para modelos de Machine Learning preditivos.

### O que ele faz?

1. **Migra dados** do banco legado (`reports`, `crimes`, `neighborhoods`)
2. **Transforma** em formato otimizado para ML com PostGIS
3. **Enriquece** com dados externos (clima, eventos, feriados)
4. **Gera features** temporais e espaciais (lags, rolling windows)
5. **Valida** qualidade dos dados processados

### Por que usar?

- ✅ **Migrations Automáticas**: Schemas criados automaticamente ao acessar a rota
- ✅ **Idempotente**: Pode executar múltiplas vezes com segurança
- ✅ **Versionado**: Todo o schema é código versionado no Git
- ✅ **Testável**: Pipeline CI/CD testa migrations em cada commit
- ✅ **Reproduzível**: Novo dev só precisa clonar o repo

---

## ⚡ Características

### 🗄️ Arquitetura de Dois Bancos

```
📦 Source DB (Legado)          →      🎯 Target DB (KB para IA)
├─ reports                     →      ├─ curated.incidents
├─ crimes                      →      ├─ curated.cells
└─ neighborhoods               →      ├─ external.weather/holidays/events
                                      ├─ features.cell_hourly
                                      └─ analytics.quality_reports
```

### 📊 Schemas Especializados

| Schema | Propósito | Tabelas |
|--------|-----------|---------|
| `curated` | Dados processados | `incidents`, `cells` |
| `external` | Dados externos | `weather`, `holidays`, `events` |
| `features` | Features para ML | `cell_hourly` |
| `analytics` | Metadados | `quality_reports`, `pipeline_logs` |

### 🚀 Pipeline CI/CD com Stages Visíveis

```
VALIDATE → BUILD → TEST → MIGRATE → DEPLOY → NOTIFY
   ✓         ✓       ✓       ✓         ✓        ✓
```

Ver [GitHub Actions workflow](.github/workflows/knowledge-base-pipeline.yml) para detalhes.

---

## 🔧 Pré-requisitos

### Software Necessário

- **Go 1.21+** ([Download](https://golang.org/dl/))
- **PostgreSQL 15+** ([Download](https://www.postgresql.org/download/))
- **PostGIS 3.4+** (extensão geoespacial)

### Verificar Instalação

```bash
go version        # go version go1.21 ou superior
psql --version    # psql (PostgreSQL) 15.0 ou superior
```

### Instalar PostGIS

#### Ubuntu/Debian
```bash
sudo apt-get update
sudo apt-get install postgresql-15-postgis-3
```

#### MacOS
```bash
brew install postgis
```

#### Verificar
```sql
-- No psql:
CREATE EXTENSION IF NOT EXISTS postgis;
SELECT PostGIS_version();
```

---

## ⚡ Configuração Rápida

### 1. Clonar Repositório

```bash
git clone <seu-repo>
cd TccRadarCampinas
```

### 2. Instalar Dependências Go

```bash
go mod download
```

### 3. Configurar Variáveis de Ambiente

Criar arquivo `.env` na raiz:

```env
# Source Database (Legado)
SOURCE_DB_HOST=localhost
SOURCE_DB_PORT=5432
SOURCE_DB_USER=postgres
SOURCE_DB_PASSWORD=sua_senha
SOURCE_DB_NAME=source_db
SOURCE_DB_SSL_MODE=disable

# Target Database (KB)
TARGET_DB_HOST=localhost
TARGET_DB_PORT=5432
TARGET_DB_USER=postgres
TARGET_DB_PASSWORD=sua_senha
TARGET_DB_NAME=radar_campinas
TARGET_DB_SSL_MODE=disable
```

### 4. Criar Bancos de Dados

```bash
# Criar bancos
createdb source_db
createdb radar_campinas

# Habilitar PostGIS no target
psql -d radar_campinas -c "CREATE EXTENSION IF NOT EXISTS postgis;"
```

### 5. Popular Source DB com Dados de Exemplo (Opcional)

```sql
-- No banco source_db, criar tabelas legadas
psql -d source_db

CREATE TABLE neighborhoods (
    neighborhood_id SERIAL PRIMARY KEY,
    name VARCHAR(100),
    latitude VARCHAR(20),
    longitude VARCHAR(20),
    neighborhood_weight INT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE crimes (
    crime_id SERIAL PRIMARY KEY,
    crime_name VARCHAR(100),
    crime_weight INT
);

CREATE TABLE reports (
    report_id SERIAL PRIMARY KEY,
    neighborhood_id INT REFERENCES neighborhoods(neighborhood_id),
    crime_id INT REFERENCES crimes(crime_id),
    report_date VARCHAR(20),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Inserir dados de exemplo
INSERT INTO neighborhoods (name, latitude, longitude, neighborhood_weight) VALUES
    ('Centro', '-22.9035', '-47.0616', 8),
    ('Cambuí', '-22.9033', '-47.0533', 7),
    ('Taquaral', '-22.8758', '-47.0533', 6);

INSERT INTO crimes (crime_name, crime_weight) VALUES
    ('Furto', 3),
    ('Roubo', 5),
    ('Homicídio', 10);

INSERT INTO reports (neighborhood_id, crime_id, report_date) VALUES
    (1, 1, '2024-01-15'),
    (1, 2, '2024-02-20'),
    (2, 1, '2024-03-10'),
    (3, 3, '2024-04-05');
```

### 6. Executar Servidor

```bash
go run cmd/server/main.go
```

🎉 Pronto! O servidor está rodando em `http://localhost:8080`

---

## 🗄️ Configuração dos Dois Bancos

### Opção 1: Bancos Locais Separados (Recomendado para Dev)

```bash
# Criar dois bancos no mesmo PostgreSQL
createdb source_db
createdb radar_campinas
```

**Conexão:**
```env
SOURCE_DB_HOST=localhost
SOURCE_DB_PORT=5432
SOURCE_DB_NAME=source_db

TARGET_DB_HOST=localhost
TARGET_DB_PORT=5432
TARGET_DB_NAME=radar_campinas
```

### Opção 2: Bancos em Servidores Diferentes

```env
# Source em servidor remoto
SOURCE_DB_HOST=legacy-db.company.com
SOURCE_DB_PORT=5432
SOURCE_DB_NAME=production_db

# Target em servidor de IA
TARGET_DB_HOST=ml-db.company.com
TARGET_DB_PORT=5432
TARGET_DB_NAME=knowledge_base
```

### Opção 3: Docker Compose (Recomendado para Testes)

Criar `docker-compose.yml`:

```yaml
version: '3.8'

services:
  source-db:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: sourcepass
      POSTGRES_DB: source_db
    ports:
      - "5433:5432"
  
  target-db:
    image: postgis/postgis:15-3.4
    environment:
      POSTGRES_PASSWORD: targetpass
      POSTGRES_DB: radar_campinas
    ports:
      - "5432:5432"
```

```bash
docker-compose up -d
```

---

## 🚀 Como Executar

### Método 1: Via API (Recomendado)

```bash
# Iniciar servidor
go run cmd/server/main.go

# Em outro terminal, fazer request
curl -X POST http://localhost:8080/api/v1/knowledge-base/generate
```

**Com parâmetros personalizados:**

```bash
curl -X POST "http://localhost:8080/api/v1/knowledge-base/generate?days_back=180&cell_resolution=1000"
```

### Método 2: Via Script Shell

```bash
# Usar valores padrão (365 dias, resolução 500m)
./scripts/run_kb_generation.sh

# Customizar parâmetros
./scripts/run_kb_generation.sh --days-back=180 --cell-resolution=1000
```

### Método 3: Aplicar Migrations Manualmente

Se quiser apenas criar os schemas sem gerar dados:

```bash
# Usar variáveis de ambiente do .env
./scripts/apply_migrations.sh

# Ou especificar credenciais
./scripts/apply_migrations.sh \
  --host=localhost \
  --port=5432 \
  --user=postgres \
  --password=minhasenha \
  --database=radar_campinas
```

---

## 🧪 Testando a Rota

### 1. Health Check

Verificar se o sistema está saudável:

```bash
curl http://localhost:8080/api/v1/knowledge-base/health
```

**Resposta esperada:**
```json
{
  "status": "healthy",
  "timestamp": "2025-10-09T20:00:00Z",
  "checks": {
    "source_db": {
      "status": "ok",
      "message": "Source database is accessible"
    },
    "target_db": {
      "status": "ok",
      "message": "Target database is accessible and schemas exist",
      "schemas": 4
    }
  }
}
```

### 2. Verificar Status da KB

Ver estatísticas da base de conhecimento:

```bash
curl http://localhost:8080/api/v1/knowledge-base/status
```

**Resposta esperada:**
```json
{
  "timestamp": "2025-10-09T20:00:00Z",
  "incidents": {
    "count": 1500
  },
  "cells": {
    "count": 450
  },
  "features": {
    "count": 180000
  },
  "last_execution": {
    "timestamp": "2025-10-09T19:30:00Z",
    "status": "success"
  },
  "quality_metrics": "{\"spatial_coverage\":0.85,\"temporal_coverage\":0.92}"
}
```

### 3. Gerar Base de Conhecimento

```bash
curl -X POST http://localhost:8080/api/v1/knowledge-base/generate
```

**Resposta esperada:**
```json
{
  "status": "success",
  "message": "Base de conhecimento gerada com sucesso",
  "elapsed_time": "1m45s",
  "cell_resolution": 500,
  "days_processed": 365,
  "start_date": "2024-10-09",
  "end_date": "2025-10-09"
}
```

### 4. Verificar Schemas no Banco

```bash
psql -d radar_campinas -c "\dn"
```

**Saída esperada:**
```
       List of schemas
    Name     |  Owner
-------------+----------
 analytics   | postgres
 curated     | postgres
 external    | postgres
 features    | postgres
 public      | postgres
```

### 5. Verificar Dados Gerados

```sql
-- Ver incidentes
SELECT COUNT(*) FROM curated.incidents;

-- Ver células
SELECT COUNT(*) FROM curated.cells;

-- Ver features
SELECT COUNT(*) FROM features.cell_hourly;

-- Ver métricas de qualidade
SELECT * FROM analytics.quality_reports ORDER BY report_date DESC LIMIT 1;
```

---

## 👁️ Visualizando o Pipeline no GitHub

### Como Acessar

1. Vá para o repositório no GitHub
2. Clique na aba **Actions**
3. Selecione o workflow **🔮 Knowledge Base Pipeline**

### Stages Visíveis

O pipeline mostra 6 stages sequenciais:

```
┌─────────────┐   ┌─────────────┐   ┌─────────────┐
│  VALIDATE   │──▶│    BUILD    │──▶│    TEST     │
│   3 jobs    │   │   1 job     │   │   2 jobs    │
└─────────────┘   └─────────────┘   └─────────────┘
                                           │
┌─────────────┐   ┌─────────────┐         │
│   NOTIFY    │◀──│   DEPLOY    │◀────────┘
│   2 jobs    │   │   1 job     │   
└─────────────┘   └─────────────┘
        ▲                 │
        │      ┌─────────────┐
        └──────│   MIGRATE   │
               │   1 job     │
               └─────────────┘
```

### Gatilhos de Execução

- **Push**: Em commits para `main` ou `develop` (apenas testes)
- **Schedule**: Toda segunda-feira às 3h (execução completa)
- **Manual**: Via botão "Run workflow" no GitHub

### Executar Manualmente

1. Vá para **Actions** → **🔮 Knowledge Base Pipeline**
2. Clique em **Run workflow**
3. Configure parâmetros:
   - `days_back`: 365 (padrão)
   - `cell_resolution`: 500 (padrão)
4. Clique em **Run workflow**

---

## 📁 Estrutura do Projeto

```
TccRadarCampinas/
├─ .github/
│  └─ workflows/
│     └─ knowledge-base-pipeline.yml    # Pipeline CI/CD
├─ cmd/
│  └─ server/
│     └─ main.go                        # Entry point
├─ internal/
│  ├─ controllers/
│  │  └─ knowledge_base_controller.go   # API handlers
│  ├─ database/
│  │  └─ migrations/
│  │     └─ knowledge_base_schema.sql   # Migrations SQL
│  └─ services/
│     └─ knowledge_base_generator.go    # Lógica de geração
├─ scripts/
│  ├─ run_kb_generation.sh              # Script para gerar KB
│  └─ apply_migrations.sh               # Script para migrations
├─ docs/
│  └─ ARCHITECTURE.md                   # Documentação detalhada
├─ go.mod
├─ go.sum
├─ .env                                 # Configuração (não committar!)
└─ README.md                            # Este arquivo
```

---

## 🔌 API Endpoints

### POST `/api/v1/knowledge-base/generate`

Gera a base de conhecimento completa.

**Query Parameters:**
- `days_back` (int): Dias para processar (padrão: 365)
- `cell_resolution` (int): 500 ou 1000 metros (padrão: 500)

**Exemplo:**
```bash
curl -X POST "http://localhost:8080/api/v1/knowledge-base/generate?days_back=180&cell_resolution=1000"
```

### GET `/api/v1/knowledge-base/health`

Verifica saúde do sistema.

**Exemplo:**
```bash
curl http://localhost:8080/api/v1/knowledge-base/health
```

### GET `/api/v1/knowledge-base/status`

Retorna estatísticas da KB.

**Exemplo:**
```bash
curl http://localhost:8080/api/v1/knowledge-base/status
```

---

## 🐛 Troubleshooting

### Erro: "relation curated.cells does not exist"

**Causa:** Migrations não foram executadas.

**Solução:**
```bash
# Opção 1: Via script
./scripts/apply_migrations.sh

# Opção 2: Via psql
psql -d radar_campinas -f internal/database/migrations/knowledge_base_schema.sql

# Opção 3: Fazer request para a rota (migrations automáticas)
curl -X POST http://localhost:8080/api/v1/knowledge-base/generate
```

### Erro: "could not connect to database"

**Causa:** Banco não está acessível.

**Verificar:**
```bash
# PostgreSQL está rodando?
pg_isready

# Credenciais corretas?
psql -h localhost -U postgres -d radar_campinas -c "SELECT 1"
```

### Performance Lenta

**Otimizações:**

1. Aumentar `BatchSize`:
```go
config := &services.KnowledgeBaseConfig{
    BatchSize: 1000, // ao invés de 500
    // ...
}
```

2. Reduzir período:
```bash
curl -X POST "http://localhost:8080/api/v1/knowledge-base/generate?days_back=90"
```

3. Usar resolução maior:
```bash
curl -X POST "http://localhost:8080/api/v1/knowledge-base/generate?cell_resolution=1000"
```

### Verificar Logs

```bash
# Ver logs do servidor
go run cmd/server/main.go

# Ver logs do pipeline no banco
psql -d radar_campinas -c "SELECT * FROM analytics.pipeline_logs ORDER BY started_at DESC LIMIT 10"
```

---

## 🤝 Contribuindo

### Como Contribuir

1. **Fork** o repositório
2. Crie uma **branch** para sua feature (`git checkout -b feature/AmazingFeature`)
3. **Commit** suas mudanças (`git commit -m 'Add some AmazingFeature'`)
4. **Push** para a branch (`git push origin feature/AmazingFeature`)
5. Abra um **Pull Request**

### Adicionando Nova Feature

Para adicionar um novo campo ou tabela:

1. Editar `internal/database/migrations/knowledge_base_schema.sql`
2. Incrementar versão (ex: v1.0.0 → v1.1.0)
3. Testar localmente
4. Abrir PR

O pipeline CI/CD validará automaticamente!

---

## 📚 Documentação Adicional

- 📖 [Arquitetura Detalhada](docs/ARCHITECTURE.md)
- 🔧 [Migrations SQL](internal/database/migrations/knowledge_base_schema.sql)
- 🚀 [GitHub Actions Pipeline](.github/workflows/knowledge-base-pipeline.yml)

---

## 📝 Licença

Este projeto está sob a licença MIT. Ver arquivo `LICENSE` para mais detalhes.

---

## 👥 Autores

**TCC Radar Campinas Team**
  * Miguel Moinhos Richena(https://github.com/MiguelRichena)
  * Aloysio Alves Ribeiro(https://github.com/AloysioLvy)
  * Diogo Lourenço Andrade(https://github.com/soothsayerdev)
---



<div align="center">

**⭐ Se este projeto foi útil, considere dar uma estrela!**

</div>
