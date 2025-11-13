-- ============================================================================
-- KNOWLEDGE BASE SCHEMA MIGRATIONS - OTIMIZADO
-- TCC Radar Campinas - Base de Conhecimento para IA Preditiva
-- ============================================================================
-- Versão: 1.1.0 (Otimizada)
-- Descrição: Schema simplificado focado apenas no essencial
-- ============================================================================

-- Habilitar extensão PostGIS
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;

-- ============================================================================
-- SCHEMA: curated
-- Propósito: Dados processados e curados de incidentes criminais
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS curated;

-- Tabela de incidentes criminais
CREATE TABLE IF NOT EXISTS curated.incidents (
    id VARCHAR(50) PRIMARY KEY,
    occurred_at TIMESTAMP NOT NULL,
    category VARCHAR(50) NOT NULL,
    severity INTEGER NOT NULL CHECK (severity BETWEEN 1 AND 10),
    geom GEOGRAPHY(POINT, 4326) NOT NULL,
    neighborhood VARCHAR(100),
    confidence FLOAT CHECK (confidence BETWEEN 0 AND 1),
    source VARCHAR(50) DEFAULT 'legacy_reports',
    cell_id VARCHAR(50),
    cell_resolution INTEGER,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Índices espaciais e temporais
CREATE INDEX IF NOT EXISTS idx_incidents_geom ON curated.incidents USING GIST(geom);
CREATE INDEX IF NOT EXISTS idx_incidents_occurred_at ON curated.incidents(occurred_at);
CREATE INDEX IF NOT EXISTS idx_incidents_cell_id ON curated.incidents(cell_id);
CREATE INDEX IF NOT EXISTS idx_incidents_category ON curated.incidents(category);
CREATE INDEX IF NOT EXISTS idx_incidents_severity ON curated.incidents(severity);

-- Tabela de células da grade espacial
CREATE TABLE IF NOT EXISTS curated.cells (
    cell_id VARCHAR(50) PRIMARY KEY,
    cell_resolution INTEGER NOT NULL,
    city VARCHAR(50) DEFAULT 'Campinas',
    geom GEOGRAPHY(POLYGON, 4326) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Índices para células
CREATE INDEX IF NOT EXISTS idx_cells_geom ON curated.cells USING GIST(geom);
CREATE INDEX IF NOT EXISTS idx_cells_resolution ON curated.cells(cell_resolution);
CREATE INDEX IF NOT EXISTS idx_cells_city ON curated.cells(city);

-- ============================================================================
-- SCHEMA: external
-- Propósito: Dados externos (apenas feriados)
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS external;

-- Tabela de feriados (criada e populada uma única vez)
CREATE TABLE IF NOT EXISTS external.holidays (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50),
    city VARCHAR(50) DEFAULT 'Campinas',
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(date, city)
);

-- Índices para holidays
CREATE INDEX IF NOT EXISTS idx_holidays_date ON external.holidays(date);
CREATE INDEX IF NOT EXISTS idx_holidays_city ON external.holidays(city);

-- Popular feriados fixos (executado apenas uma vez)
INSERT INTO external.holidays (date, name, type, city) VALUES
    -- Feriados Nacionais 2024
    ('2024-01-01', 'Ano Novo', 'nacional', 'Campinas'),
    ('2024-02-13', 'Carnaval', 'nacional', 'Campinas'),
    ('2024-04-21', 'Tiradentes', 'nacional', 'Campinas'),
    ('2024-05-01', 'Dia do Trabalho', 'nacional', 'Campinas'),
    ('2024-05-30', 'Corpus Christi', 'nacional', 'Campinas'),
    ('2024-07-11', 'Fundação de Campinas', 'municipal', 'Campinas'),
    ('2024-09-07', 'Independência do Brasil', 'nacional', 'Campinas'),
    ('2024-10-12', 'Nossa Senhora Aparecida', 'nacional', 'Campinas'),
    ('2024-11-02', 'Finados', 'nacional', 'Campinas'),
    ('2024-11-15', 'Proclamação da República', 'nacional', 'Campinas'),
    ('2024-11-20', 'Consciência Negra', 'nacional', 'Campinas'),
    ('2024-12-25', 'Natal', 'nacional', 'Campinas'),
    
    -- Feriados Nacionais 2025
    ('2025-01-01', 'Ano Novo', 'nacional', 'Campinas'),
    ('2025-03-04', 'Carnaval', 'nacional', 'Campinas'),
    ('2025-04-21', 'Tiradentes', 'nacional', 'Campinas'),
    ('2025-05-01', 'Dia do Trabalho', 'nacional', 'Campinas'),
    ('2025-06-19', 'Corpus Christi', 'nacional', 'Campinas'),
    ('2025-07-11', 'Fundação de Campinas', 'municipal', 'Campinas'),
    ('2025-09-07', 'Independência do Brasil', 'nacional', 'Campinas'),
    ('2025-10-12', 'Nossa Senhora Aparecida', 'nacional', 'Campinas'),
    ('2025-11-02', 'Finados', 'nacional', 'Campinas'),
    ('2025-11-15', 'Proclamação da República', 'nacional', 'Campinas'),
    ('2025-11-20', 'Consciência Negra', 'nacional', 'Campinas'),
    ('2025-12-25', 'Natal', 'nacional', 'Campinas'),
    
    -- Feriados 2026
    ('2026-01-01', 'Ano Novo', 'nacional', 'Campinas'),
    ('2026-02-17', 'Carnaval', 'nacional', 'Campinas'),
    ('2026-04-21', 'Tiradentes', 'nacional', 'Campinas'),
    ('2026-05-01', 'Dia do Trabalho', 'nacional', 'Campinas'),
    ('2026-06-04', 'Corpus Christi', 'nacional', 'Campinas'),
    ('2026-07-11', 'Fundação de Campinas', 'municipal', 'Campinas'),
    ('2026-09-07', 'Independência do Brasil', 'nacional', 'Campinas'),
    ('2026-10-12', 'Nossa Senhora Aparecida', 'nacional', 'Campinas'),
    ('2026-11-02', 'Finados', 'nacional', 'Campinas'),
    ('2026-11-15', 'Proclamação da República', 'nacional', 'Campinas'),
    ('2026-11-20', 'Consciência Negra', 'nacional', 'Campinas'),
    ('2026-12-25', 'Natal', 'nacional', 'Campinas')
ON CONFLICT (date, city) DO NOTHING;

-- ============================================================================
-- SCHEMA: features
-- Propósito: Features engenheiradas para modelos de ML (SIMPLIFICADO)
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS features;

-- Tabela de features por célula e hora (SEM dados de clima e eventos)
CREATE TABLE IF NOT EXISTS features.cell_hourly (
    id SERIAL PRIMARY KEY,
    cell_id VARCHAR(50) NOT NULL,
    ts TIMESTAMP NOT NULL,
    
    -- Target variable
    y_count INTEGER DEFAULT 0,
    
    -- Lag features
    lag_1h INTEGER DEFAULT 0,
    lag_24h INTEGER DEFAULT 0,
    lag_7d INTEGER DEFAULT 0,
    
    -- Rolling window features
    roll_3h_sum INTEGER DEFAULT 0,
    roll_24h_sum INTEGER DEFAULT 0,
    roll_7d_sum INTEGER DEFAULT 0,
    roll_7d_avg FLOAT,
    roll_7d_std FLOAT,
    
    -- Temporal features
    dow INTEGER, -- day of week (0-6)
    hour INTEGER, -- hour of day (0-23)
    is_weekend BOOLEAN,
    is_business_hours BOOLEAN,
    
    -- Calendar features
    holiday BOOLEAN DEFAULT FALSE,
    day_before_holiday BOOLEAN DEFAULT FALSE,
    day_after_holiday BOOLEAN DEFAULT FALSE,
    
    -- Spatial features
    neighbor_avg_crime FLOAT,
    
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(cell_id, ts)
);

-- Índices para features
CREATE INDEX IF NOT EXISTS idx_features_cell_id ON features.cell_hourly(cell_id);
CREATE INDEX IF NOT EXISTS idx_features_ts ON features.cell_hourly(ts);
CREATE INDEX IF NOT EXISTS idx_features_cell_ts ON features.cell_hourly(cell_id, ts);
CREATE INDEX IF NOT EXISTS idx_features_dow ON features.cell_hourly(dow);
CREATE INDEX IF NOT EXISTS idx_features_hour ON features.cell_hourly(hour);

-- ============================================================================
-- SCHEMA: analytics
-- Propósito: Metadados e métricas de qualidade
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS analytics;

-- Tabela de relatórios de qualidade
CREATE TABLE IF NOT EXISTS analytics.quality_reports (
    id SERIAL PRIMARY KEY,
    report_date DATE UNIQUE NOT NULL,
    metrics JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_quality_reports_date ON analytics.quality_reports(report_date);
CREATE INDEX IF NOT EXISTS idx_quality_reports_metrics ON analytics.quality_reports USING GIN(metrics);

-- Tabela de logs de execução do pipeline
CREATE TABLE IF NOT EXISTS analytics.pipeline_logs (
    id SERIAL PRIMARY KEY,
    execution_id UUID UNIQUE NOT NULL,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP,
    status VARCHAR(20),
    phase VARCHAR(50),
    records_processed INTEGER,
    error_message TEXT,
    execution_time_seconds INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Índices
CREATE INDEX IF NOT EXISTS idx_pipeline_logs_execution_id ON analytics.pipeline_logs(execution_id);
CREATE INDEX IF NOT EXISTS idx_pipeline_logs_started_at ON analytics.pipeline_logs(started_at);
CREATE INDEX IF NOT EXISTS idx_pipeline_logs_status ON analytics.pipeline_logs(status);

-- ============================================================================
-- VIEWS ÚTEIS
-- ============================================================================

-- View para estatísticas de incidentes por célula
CREATE OR REPLACE VIEW analytics.cell_statistics AS
SELECT 
    cell_id,
    COUNT(*) as total_incidents,
    COUNT(DISTINCT DATE(occurred_at)) as days_with_incidents,
    AVG(severity) as avg_severity,
    MIN(occurred_at) as first_incident,
    MAX(occurred_at) as last_incident
FROM curated.incidents
WHERE cell_id IS NOT NULL
GROUP BY cell_id;

-- View para hotspots de criminalidade
CREATE OR REPLACE VIEW analytics.crime_hotspots AS
SELECT 
    c.cell_id,
    c.geom,
    COUNT(i.id) as incident_count,
    AVG(i.severity) as avg_severity,
    MAX(i.occurred_at) as last_incident
FROM curated.cells c
LEFT JOIN curated.incidents i ON c.cell_id = i.cell_id
GROUP BY c.cell_id, c.geom
HAVING COUNT(i.id) > 0
ORDER BY incident_count DESC;

-- View para cobertura temporal
CREATE OR REPLACE VIEW analytics.temporal_coverage AS
SELECT 
    DATE(occurred_at) as date,
    COUNT(*) as incidents,
    COUNT(DISTINCT cell_id) as cells_affected,
    AVG(severity) as avg_severity
FROM curated.incidents
GROUP BY DATE(occurred_at)
ORDER BY date;

-- ============================================================================
-- FUNCTIONS ÚTEIS
-- ============================================================================

-- Função trigger para atualizar updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Aplicar triggers
DROP TRIGGER IF EXISTS update_incidents_updated_at ON curated.incidents;
CREATE TRIGGER update_incidents_updated_at
    BEFORE UPDATE ON curated.incidents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_features_updated_at ON features.cell_hourly;
CREATE TRIGGER update_features_updated_at
    BEFORE UPDATE ON features.cell_hourly
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_quality_reports_updated_at ON analytics.quality_reports;
CREATE TRIGGER update_quality_reports_updated_at
    BEFORE UPDATE ON analytics.quality_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- COMENTÁRIOS
-- ============================================================================

COMMENT ON SCHEMA curated IS 'Dados processados e curados de incidentes criminais';
COMMENT ON SCHEMA external IS 'Dados externos (feriados)';
COMMENT ON SCHEMA features IS 'Features engenheiradas para modelos de ML';
COMMENT ON SCHEMA analytics IS 'Metadados e métricas de qualidade';

-- ============================================================================
-- REGISTRO DE MIGRAÇÃO
-- ============================================================================

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = 'schema_migrations') THEN
        CREATE TABLE public.schema_migrations (
            version VARCHAR(50) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT NOW()
        );
    END IF;
    
    INSERT INTO public.schema_migrations (version, applied_at)
    VALUES ('knowledge_base_v1.1.0_optimized', NOW())
    ON CONFLICT (version) DO NOTHING;
END $$;

-- Log de sucesso
DO $$
BEGIN
    RAISE NOTICE '✅ Knowledge Base Schema Migrations (OTIMIZADO) aplicadas com sucesso!';
    RAISE NOTICE '📊 Schemas criados: curated, external, features, analytics';
    RAISE NOTICE '🗂️  Tabelas: incidents, cells, holidays, cell_hourly, quality_reports, pipeline_logs';
    RAISE NOTICE '⚡ Otimizações: Removido weather, events; Feriados pré-populados';
    RAISE NOTICE '🚀 Sistema pronto!';
END $$;