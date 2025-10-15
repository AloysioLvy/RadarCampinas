# 🏗️ Arquitetura da Base de Conhecimento - Radar Campinas

## 📋 Índice

1. [Visão Geral](#visão-geral)
2. [Decisão: Migrations vs Banco Fixo](#decisão-migrations-vs-banco-fixo)
3. [Arquitetura de Dois Bancos](#arquitetura-de-dois-bancos)
4. [Schemas e Tabelas](#schemas-e-tabelas)
5. [Fluxo de Geração da KB](#fluxo-de-geração-da-kb)
6. [Pipeline de CI/CD](#pipeline-de-cicd)
7. [Execução Manual](#execução-manual)
8. [Monitoramento e Qualidade](#monitoramento-e-qualidade)

---

## 🎯 Visão Geral

O sistema **Radar Campinas** é uma plataforma de análise preditiva de criminalidade que utiliza machine learning para prever padrões criminais na cidade de Campinas/SP. A **Base de Conhecimento** (KB - Knowledge Base) é o coração do sistema, agregando dados históricos, externos e features engenheiradas para alimentar os modelos de IA.

### Objetivos da KB

- **Consolidar** dados históricos de crimes do banco legado
- **Enriquecer** com dados externos (clima, eventos, feriados)
- **Gerar features** temporais e espaciais para ML
- **Validar** qualidade dos dados processados
- **Facilitar** acesso aos dados para treinamento de modelos

---

## 🤔 Decisão: Migrations vs Banco Fixo

### ✅ **Recomendação: Migrations Automáticas (Escolha Implementada)**

Decidimos implementar **migrations automáticas** que são executadas sempre que a rota `/api/v1/knowledge-base/generate` é acessada. Esta decisão foi tomada após análise cuidadosa dos prós e contras de cada abordagem.

### 📊 Comparação das Abordagens

| Aspecto | 🏗️ Migrations Automáticas | 🗄️ Banco Fixo Pré-configurado |
|---------|---------------------------|-------------------------------|
| **Versionamento** | ✅ Versionado em Git | ❌ Difícil versionar estrutura |
| **Reprodutibilidade** | ✅ Reproduzível em qualquer ambiente | ⚠️ Requer setup manual |
| **CI/CD** | ✅ Testável em pipeline | ⚠️ Requer DB de staging |
| **Onboarding** | ✅ Novo dev só clona repo | ❌ Precisa receber dump/instruções |
| **Rollback** | ✅ Git revert + re-run | ❌ Backup/restore manual |
| **Evolução** | ✅ Adicionar campos é simples | ⚠️ Precisa coordenar mudanças |
| **Drift** | ✅ Migrations garantem consistência | ❌ DBs podem divergir |
| **Performance inicial** | ⚠️ 1-2s na primeira execução | ✅ Instantâneo |
| **Documentação** | ✅ Schema é código | ⚠️ Documentação externa |

### 🎯 Por que Migrations Venceu?

#### **1. Versionamento e Git**
```sql
-- Migrations são código! Posso fazer:
git log internal/database/migrations/
git diff HEAD~1 knowledge_base_schema.sql
git blame knowledge_base_schema.sql
```

#### **2. Reprodutibilidade Total**
```bash
# Novo desenvolvedor:
git clone <repo>
go run cmd/server/main.go  # Migrations aplicadas automaticamente!
```

#### **3. Idempotência**
```go
// Migrations são idempotentes - pode rodar N vezes
CREATE TABLE IF NOT EXISTS curated.incidents ...
CREATE SCHEMA IF NOT EXISTS external ...
```

#### **4. Testabilidade em CI/CD**
```yaml
# GitHub Actions testa migrations em cada commit
test-migrations:
  - Apply migrations on fresh Postgres
  - Verify schemas created
  - Test idempotency
```

#### **5. Evolução Incremental**
```sql
-- v1.0.0: Schema inicial
CREATE TABLE curated.incidents (...)

-- v1.1.0: Adicionar campo (futuro)
ALTER TABLE curated.incidents ADD COLUMN IF NOT EXISTS risk_score FLOAT;
```

### ⚠️ Desvantagens do Banco Fixo (Por que NÃO escolhemos)

#### **1. Schema Drift**
```
Dev 1: Adiciona coluna localmente
Dev 2: Trabalha com schema antigo
Produção: Schema diferente dos dois!
❌ Disaster waiting to happen
```

#### **2. Onboarding Complexo**
```
Novo dev: Como setup o banco?
Senior: Baixa esse dump de 2GB...
         Ou roda esse script...
         Mas cuidado com os dados sensíveis...
❌ Friction desnecessária
```

#### **3. Difficult Rollbacks**
```
Bug no schema novo?
Banco fixo: pg_dump, pg_restore, rezar...
Migrations: git revert + re-run
✅ Clean & predictable
```

### 💡 Quando Banco Fixo Faria Sentido?

Banco fixo seria válido apenas se:
- Sistema 100% estável, zero mudanças no schema
- Dados sensíveis que não podem ser recriados
- Performance crítica (mas migrations levam ~1-2s)
- Equipe muito pequena (1-2 devs)

**Para nosso caso**: Estamos em desenvolvimento ativo, evoluindo o schema, com CI/CD, então migrations são a escolha óbvia! 🎯

---

## 🗄️ Arquitetura de Dois Bancos

### Diagrama

```
┌─────────────────────────────────────────────────────────────────┐
│                    🏢 SOURCE DATABASE (Legado)                   │
│                                                                   │
│  ┌─────────────┐    ┌──────────────┐    ┌──────────────────┐   │
│  │   reports   │    │ crimes       │    │ neighborhoods    │   │
│  │             │    │              │    │                  │   │
│  │ - report_id │───▶│ - crime_id   │    │ - neighborhood_id│   │
│  │ - crime_id  │    │ - crime_name │    │ - name           │   │
│  │ - neigh_id  │───▶│ - weight     │◀───│ - lat, lon       │   │
│  │ - date      │    └──────────────┘    │ - weight         │   │
│  └─────────────┘                        └──────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ ETL Process
                              │ (KnowledgeBaseGenerator)
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                  🎯 TARGET DATABASE (KB para IA)                 │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ 📊 SCHEMA: curated (Dados Processados)                   │   │
│  │  ├─ incidents    : Crimes transformados + enriquecidos   │   │
│  │  └─ cells        : Grade espacial (500m ou 1000m)        │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ 🌍 SCHEMA: external (Dados Externos)                     │   │
│  │  ├─ weather      : Clima (temperatura, chuva, etc)       │   │
│  │  ├─ holidays     : Feriados nacionais/municipais         │   │
│  │  └─ events       : Shows, jogos, manifestações           │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ ⚙️ SCHEMA: features (ML Features)                         │   │
│  │  └─ cell_hourly  : Features por célula e hora            │   │
│  │                    (lags, rolling windows, temporais)     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ 📈 SCHEMA: analytics (Metadados)                          │   │
│  │  ├─ quality_reports : Métricas de qualidade              │   │
│  │  └─ pipeline_logs   : Logs de execução                   │   │
│  └──────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### Por que Dois Bancos?

#### **Source DB (Legado)**
- **Propósito**: Dados operacionais do sistema antigo
- **Schema**: Não controlamos, já existe
- **Tecnologia**: PostgreSQL padrão
- **Conteúdo**: `reports`, `crimes`, `neighborhoods`
- **Acesso**: Apenas leitura (no KB Generator)

#### **Target DB (KB)**
- **Propósito**: Base otimizada para IA preditiva
- **Schema**: Controlamos 100%, desenhado para ML
- **Tecnologia**: PostgreSQL + PostGIS (geoespacial)
- **Conteúdo**: 4 schemas especializados
- **Acesso**: Leitura/escrita (KB Generator + modelos ML)

### Benefícios da Separação

1. **Isolamento**: Sistema legado não é afetado
2. **Performance**: Índices otimizados para queries de ML
3. **Escalabilidade**: Podemos escalar KB independentemente
4. **Segurança**: Credenciais separadas
5. **Flexibilidade**: Podemos adicionar outros sources no futuro

---

## 📊 Schemas e Tabelas

### 1️⃣ Schema: `curated`
**Propósito**: Dados processados e curados de crimes

#### Tabela: `curated.incidents`
```sql
CREATE TABLE curated.incidents (
    id VARCHAR(50) PRIMARY KEY,            -- rpt_123
    occurred_at TIMESTAMP NOT NULL,        -- Quando ocorreu
    category VARCHAR(50) NOT NULL,         -- Hediondo / Comum
    severity INTEGER (1-10),               -- Gravidade
    geom GEOGRAPHY(POINT, 4326),           -- PostGIS point
    neighborhood VARCHAR(100),             -- Bairro
    confidence FLOAT (0-1),                -- Score de confiança
    source VARCHAR(50),                    -- legacy_reports
    cell_id VARCHAR(50),                   -- CAMP-500-1234
    cell_resolution INTEGER,               -- 500 ou 1000
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

#### Tabela: `curated.cells`
```sql
CREATE TABLE curated.cells (
    cell_id VARCHAR(50) PRIMARY KEY,       -- CAMP-500-1234
    cell_resolution INTEGER NOT NULL,      -- 500 ou 1000 metros
    city VARCHAR(50),                      -- Campinas
    geom GEOGRAPHY(POLYGON, 4326),         -- Polígono da célula
    created_at TIMESTAMP
);
```

**Índices Espaciais**: GiST indexes para queries geoespaciais rápidas

---

### 2️⃣ Schema: `external`
**Propósito**: Dados externos que influenciam criminalidade

#### Tabela: `external.weather`
```sql
CREATE TABLE external.weather (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    rain_mm FLOAT,                         -- Precipitação
    temp_c FLOAT,                          -- Temperatura
    humidity FLOAT,                        -- Umidade
    wind_speed FLOAT,                      -- Vento
    pressure FLOAT,                        -- Pressão atmosférica
    city VARCHAR(50),
    source VARCHAR(50),
    created_at TIMESTAMP,
    UNIQUE(timestamp, city)
);
```

#### Tabela: `external.holidays`
```sql
CREATE TABLE external.holidays (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    name VARCHAR(100) NOT NULL,            -- Natal, Tiradentes, etc
    type VARCHAR(50),                      -- nacional, estadual, municipal
    city VARCHAR(50),
    created_at TIMESTAMP,
    UNIQUE(date, city)
);
```

#### Tabela: `external.events`
```sql
CREATE TABLE external.events (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    name VARCHAR(200) NOT NULL,            -- Show no Estádio
    geom GEOGRAPHY(POINT, 4326),           -- Localização
    attendance INTEGER,                    -- Público estimado
    type VARCHAR(50),                      -- show, esporte, feira
    impact_radius INTEGER,                 -- 1000 metros
    city VARCHAR(50),
    source VARCHAR(50),
    created_at TIMESTAMP,
    UNIQUE(timestamp, name, city)
);
```

**Por que isso importa?**: Crimes aumentam perto de eventos, em dias chuvosos, em feriados, etc.

---

### 3️⃣ Schema: `features`
**Propósito**: Features engenheiradas para modelos de ML

#### Tabela: `features.cell_hourly`
```sql
CREATE TABLE features.cell_hourly (
    id SERIAL PRIMARY KEY,
    cell_id VARCHAR(50) NOT NULL,
    ts TIMESTAMP NOT NULL,
    
    -- Target variable
    y_count INTEGER DEFAULT 0,             -- Crimes nesta hora
    
    -- Lag features (valores passados)
    lag_1h INTEGER DEFAULT 0,              -- Crimes 1h atrás
    lag_24h INTEGER DEFAULT 0,             -- Crimes 24h atrás
    lag_7d INTEGER DEFAULT 0,              -- Crimes 7 dias atrás
    
    -- Rolling window features
    roll_3h_sum INTEGER DEFAULT 0,         -- Soma últimas 3h
    roll_24h_sum INTEGER DEFAULT 0,        -- Soma últimas 24h
    roll_7d_sum INTEGER DEFAULT 0,         -- Soma últimos 7 dias
    roll_7d_avg FLOAT,                     -- Média últimos 7 dias
    roll_7d_std FLOAT,                     -- Desvio padrão
    
    -- Temporal features
    dow INTEGER,                           -- Dia da semana (0-6)
    hour INTEGER,                          -- Hora do dia (0-23)
    is_weekend BOOLEAN,                    -- É fim de semana?
    is_business_hours BOOLEAN,             -- Horário comercial?
    
    -- Weather features
    weather_rain_mm FLOAT,
    weather_temp_c FLOAT,
    weather_humidity FLOAT,
    
    -- Calendar features
    holiday BOOLEAN DEFAULT FALSE,
    day_before_holiday BOOLEAN DEFAULT FALSE,
    day_after_holiday BOOLEAN DEFAULT FALSE,
    
    -- Event features
    nearby_events INTEGER DEFAULT 0,
    event_attendance INTEGER DEFAULT 0,
    
    -- Spatial features
    neighbor_avg_crime FLOAT,              -- Média dos vizinhos
    
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    UNIQUE(cell_id, ts)
);
```

**Por que tantas features?**: Modelos de ML precisam de contexto temporal e espacial para fazer boas predições.

---

### 4️⃣ Schema: `analytics`
**Propósito**: Metadados e monitoramento

#### Tabela: `analytics.quality_reports`
```sql
CREATE TABLE analytics.quality_reports (
    id SERIAL PRIMARY KEY,
    report_date DATE UNIQUE NOT NULL,
    metrics JSONB NOT NULL,                -- {spatial_coverage: 0.85, ...}
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

#### Tabela: `analytics.pipeline_logs`
```sql
CREATE TABLE analytics.pipeline_logs (
    id SERIAL PRIMARY KEY,
    execution_id UUID UNIQUE NOT NULL,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP,
    status VARCHAR(20),                    -- running, success, failed
    phase VARCHAR(50),                     -- migrate, spatial_grid, etc
    records_processed INTEGER,
    error_message TEXT,
    execution_time_seconds INTEGER,
    created_at TIMESTAMP
);
```

---

## 🔄 Fluxo de Geração da KB

### Sequência de Execução

```
┌─────────────────────────────────────────────────────────────────┐
│ POST /api/v1/knowledge-base/generate                            │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 🔧 FASE 0: Run Migrations                                       │
│ ├─ Verificar se migrations já aplicadas (idempotente)           │
│ ├─ Executar knowledge_base_schema.sql                           │
│ ├─ Criar schemas: curated, external, features, analytics        │
│ ├─ Criar tabelas + índices + views + functions                  │
│ └─ Log: analytics.pipeline_logs                                 │
│ ⏱️  Tempo: ~1-2 segundos (primeira vez), ~0.1s (subsequentes)   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 📊 FASE 1: Migrate Historical Data                              │
│ ├─ Query: SELECT reports JOIN neighborhoods JOIN crimes         │
│ ├─ Transform: report → incident (mapear categorias)             │
│ ├─ Validate: coordenadas dentro de Campinas                     │
│ ├─ Batch insert: curated.incidents (500-1000 registros/lote)    │
│ └─ Log: registros processados + ignorados                       │
│ ⏱️  Tempo: ~10-30 segundos (10k registros)                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 🗺️  FASE 2: Generate Spatial Grid                               │
│ ├─ Calcular células para bounding box de Campinas               │
│ ├─ Resolução: 500m ou 1000m (configurável)                      │
│ ├─ Gerar polígonos: ST_MakeEnvelope(lon, lat, lon+δ, lat+δ)    │
│ ├─ Batch insert: curated.cells (~200-800 células)               │
│ └─ Log: células geradas                                         │
│ ⏱️  Tempo: ~2-5 segundos                                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 🎯 FASE 3: Assign Cells to Incidents                            │
│ ├─ Spatial join: ST_Contains(cell.geom, incident.geom)          │
│ ├─ Update: incidents SET cell_id = cells.cell_id                │
│ └─ Log: incidentes atribuídos                                   │
│ ⏱️  Tempo: ~5-10 segundos (10k registros)                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ 🌦️  FASE 4: Ingest External Data                                │
│ ├─ Weather: Inserir dados de clima (API ou mock)                │
│ ├─ Holidays: Inserir calendário de feriados 2025                │
│ ├─ Events: Inserir eventos relevantes (shows, jogos)            │
│ └─ Log: total de registros externos                             │
│ ⏱️  Tempo: ~1-3 segundos (dados mock)                           │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ ⚙️  FASE 5: Generate Temporal Features                          │
│ ├─ Para cada hora no período (StartDate → EndDate):             │
│ │   ├─ Calcular y_count (crimes nesta hora)                     │
│ │   ├─ Calcular lags (1h, 24h, 7d atrás)                        │
│ │   ├─ Calcular rolling windows (3h, 24h, 7d)                   │
│ │   ├─ Adicionar features temporais (dow, hour, weekend)        │
│ │   ├─ Join com weather                                         │
│ │   ├─ Join com holidays                                        │
│ │   └─ INSERT/UPDATE features.cell_hourly                       │
│ └─ Log: horas processadas                                       │
│ ⏱️  Tempo: ~30-120 segundos (1 ano de dados)                    │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ ✓ FASE 6: Validate Data Quality                                │
│ ├─ Calcular métricas:                                           │
│ │   ├─ Cobertura espacial: células com dados / total células    │
│ │   ├─ Cobertura temporal: horas com dados / total horas        │
│ │   ├─ Taxa de duplicação: duplicados / total                   │
│ │   └─ Completude de features: % campos preenchidos             │
│ ├─ Inserir analytics.quality_reports                            │
│ ├─ Validações críticas:                                         │
│ │   ├─ Cobertura espacial > 10%? ✓                              │
│ │   └─ Taxa duplicação < 50%? ✓                                 │
│ └─ Log: métricas de qualidade                                   │
│ ⏱️  Tempo: ~3-5 segundos                                         │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│ ✅ SUCESSO! Base de conhecimento gerada                         │
│ ⏱️  Tempo total: ~1-3 minutos (1 ano de dados)                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🚀 Pipeline de CI/CD

### Visualização no GitHub Actions

O pipeline é organizado em **6 stages sequenciais** com dependências claras:

```
STAGE 1: VALIDATE          STAGE 2: BUILD          STAGE 3: TEST
┌──────────────┐            ┌──────────────┐        ┌──────────────┐
│ Lint Go      │───┐        │              │        │ Unit Tests   │
└──────────────┘   │        │              │        └──────────────┘
┌──────────────┐   ├───────▶│ Build App    │───────▶┌──────────────┐
│ Validate SQL │───┤        │              │        │ Integration  │
└──────────────┘   │        │              │        │ Tests        │
┌──────────────┐   │        └──────────────┘        └──────────────┘
│ Security Scan│───┘                                        │
└──────────────┘                                            │
                                                            ▼
STAGE 4: MIGRATE           STAGE 5: DEPLOY         STAGE 6: NOTIFY
┌──────────────┐            ┌──────────────┐        ┌──────────────┐
│              │            │              │        │ Notify       │
│ Test         │───────────▶│ Generate KB  │───────▶│ Success      │
│ Migrations   │            │ (schedule)   │        │              │
│              │            │              │        └──────────────┘
└──────────────┘            └──────────────┘
```

### Stages Detalhados

#### Stage 1: Validate (Parallel)
- **lint-go**: golangci-lint para code quality
- **validate-sql**: Verificar sintaxe SQL
- **security-scan**: Gosec para vulnerabilidades

#### Stage 2: Build
- **build-app**: Compilar binário Go
- Upload artifact para stages seguintes

#### Stage 3: Test (Parallel)
- **unit-tests**: Tests unitários + coverage
- **integration-tests**: Tests com PostgreSQL real

#### Stage 4: Migrate
- **test-migrations**: Aplicar em DB temporário
- Verificar schemas e tabelas criados
- Testar idempotência

#### Stage 5: Deploy (Conditional)
- **generate-knowledge-base**: Executar geração completa
- Apenas em schedule ou manual trigger
- Usa Docker containers para DBs

#### Stage 6: Notify
- **notify-success**: Notificação em caso de sucesso
- **notify-failure**: Notificação em caso de falha

### Gatilhos de Execução

```yaml
# 1. Schedule (1x por semana)
schedule:
  - cron: '0 3 * * 1'  # Segunda às 3h

# 2. Manual (via GitHub UI)
workflow_dispatch:
  inputs:
    days_back: 365
    cell_resolution: 500

# 3. Push (apenas testes, não gera KB)
push:
  branches: [main, develop]
  paths: ['internal/**', '.github/workflows/**']
```

---

## 🖥️ Execução Manual

### Opção 1: Via API (Recomendado)

```bash
# Fazer request para a rota
curl -X POST http://localhost:8080/api/v1/knowledge-base/generate

# Response esperado
{
  "status": "Base de conhecimento gerada com sucesso"
}
```

### Opção 2: Via Script Shell

```bash
# Executar script auxiliar
./scripts/run_kb_generation.sh

# Com parâmetros personalizados
./scripts/run_kb_generation.sh --days-back=180 --cell-resolution=1000
```

### Opção 3: Via Go Direto

```bash
# Compilar e executar
go build -o radar-kb ./cmd/server
./radar-kb generate-kb
```

### Opção 4: Docker Compose

```bash
# Subir DBs + aplicação
docker-compose up -d

# Executar geração
docker-compose exec app /app/radar-kb generate-kb
```

---

## 📊 Monitoramento e Qualidade

### Métricas Coletadas

```json
{
  "spatial_coverage": 0.85,      // 85% das células têm dados
  "temporal_coverage": 0.92,     // 92% das horas têm dados
  "duplication_rate": 0.03,      // 3% de duplicação
  "feature_completeness": 0.98   // 98% das features preenchidas
}
```

### Queries Úteis

```sql
-- Ver últimas execuções do pipeline
SELECT 
    execution_id,
    started_at,
    finished_at,
    status,
    phase,
    records_processed
FROM analytics.pipeline_logs
ORDER BY started_at DESC
LIMIT 10;

-- Ver células com mais crimes
SELECT 
    cell_id,
    COUNT(*) as crime_count,
    AVG(severity) as avg_severity
FROM curated.incidents
WHERE occurred_at >= NOW() - INTERVAL '30 days'
GROUP BY cell_id
ORDER BY crime_count DESC
LIMIT 10;

-- Ver features de uma célula específica
SELECT *
FROM features.cell_hourly
WHERE cell_id = 'CAMP-500-1234'
ORDER BY ts DESC
LIMIT 24;  -- últimas 24 horas
```

### Alertas

O sistema valida automaticamente:
- ✅ Cobertura espacial > 10%
- ✅ Taxa de duplicação < 50%
- ⚠️ Se falhar, erro é retornado

---

## 🔍 Troubleshooting

### Erro: "relation curated.cells does not exist"

**Causa**: Migrations não foram executadas

**Solução**:
```bash
# 1. Verificar se arquivo existe
ls -la internal/database/migrations/knowledge_base_schema.sql

# 2. Aplicar manualmente
psql -h localhost -U postgres -d radar_campinas -f internal/database/migrations/knowledge_base_schema.sql

# 3. Verificar schemas
psql -h localhost -U postgres -d radar_campinas -c "\dn"
```

### Performance Lenta

**Otimizações**:
- Aumentar `BatchSize` no config (padrão: 500)
- Reduzir período de `StartDate` / `EndDate`
- Usar `cell_resolution` maior (1000m ao invés de 500m)
- Criar índices adicionais se necessário

### Dados Faltando

**Verificações**:
```sql
-- Quantos incidentes foram migrados?
SELECT COUNT(*) FROM curated.incidents;

-- Quantas células foram geradas?
SELECT COUNT(*) FROM curated.cells;

-- Quantas features foram geradas?
SELECT COUNT(*) FROM features.cell_hourly;
```

---

## 📚 Referências

- [PostGIS Documentation](https://postgis.net/docs/)
- [PostgreSQL Migration Best Practices](https://www.postgresql.org/docs/current/ddl-schemas.html)
- [GitHub Actions CI/CD Guide](https://docs.github.com/en/actions)
- [Go Database/SQL Tutorial](https://go.dev/doc/database/querying)

---

## 🤝 Contribuindo

Para adicionar uma nova feature ou schema:

1. Editar `internal/database/migrations/knowledge_base_schema.sql`
2. Incrementar versão (ex: v1.0.0 → v1.1.0)
3. Testar localmente
4. Abrir PR com descrição das mudanças
5. Pipeline CI/CD valida automaticamente

---

**Última atualização**: 2025-10-09  
**Versão**: 1.0.0  
**Autor**: TCC Radar Campinas 
