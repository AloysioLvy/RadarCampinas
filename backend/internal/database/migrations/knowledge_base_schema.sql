-- ============================================================================
-- KNOWLEDGE BASE SCHEMA MIGRATIONS
-- TCC Radar Campinas - Base de Conhecimento para IA Preditiva
-- ============================================================================
-- Versão: 1.0.0
-- Data: 2025-10-09
-- Descrição: Cria todos os schemas e tabelas necessários para a base de
--           conhecimento de criminalidade preditiva em Campinas
-- ============================================================================

-- Habilitar extensão PostGIS para dados geoespaciais
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS postgis_topology;

-- ============================================================================
-- SCHEMA: curated
-- Propósito: Dados processados e curados de incidentes criminais
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS curated;

-- Tabela de incidentes criminais (dados migrados do DB legado)
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

-- Índices espaciais e temporais para performance
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
-- Propósito: Dados externos que influenciam a criminalidade
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS external;

-- Tabela de dados meteorológicos
CREATE TABLE IF NOT EXISTS external.weather (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    rain_mm FLOAT,
    temp_c FLOAT,
    humidity FLOAT,
    wind_speed FLOAT,
    pressure FLOAT,
    city VARCHAR(50) DEFAULT 'Campinas',
    source VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(timestamp, city)
);

-- Índices para weather
CREATE INDEX IF NOT EXISTS idx_weather_timestamp ON external.weather(timestamp);
CREATE INDEX IF NOT EXISTS idx_weather_city ON external.weather(city);

-- Tabela de feriados e datas especiais
CREATE TABLE IF NOT EXISTS external.holidays (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50), -- nacional, estadual, municipal
    city VARCHAR(50) DEFAULT 'Campinas',
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(date, city)
);

-- Índices para holidays
CREATE INDEX IF NOT EXISTS idx_holidays_date ON external.holidays(date);
CREATE INDEX IF NOT EXISTS idx_holidays_city ON external.holidays(city);
CREATE INDEX IF NOT EXISTS idx_holidays_type ON external.holidays(type);

-- Tabela de eventos (shows, jogos, manifestações)
CREATE TABLE IF NOT EXISTS external.events (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL,
    name VARCHAR(200) NOT NULL,
    geom GEOGRAPHY(POINT, 4326) NOT NULL,
    attendance INTEGER,
    type VARCHAR(50), -- show, esporte, feira, manifestacao, etc
    impact_radius INTEGER DEFAULT 1000, -- metros
    city VARCHAR(50) DEFAULT 'Campinas',
    source VARCHAR(50),
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(timestamp, name, city)
);

-- Índices para events
CREATE INDEX IF NOT EXISTS idx_events_geom ON external.events USING GIST(geom);
CREATE INDEX IF NOT EXISTS idx_events_timestamp ON external.events(timestamp);
CREATE INDEX IF NOT EXISTS idx_events_type ON external.events(type);
CREATE INDEX IF NOT EXISTS idx_events_city ON external.events(city);

-- ============================================================================
-- SCHEMA: features
-- Propósito: Features engenheiradas para modelos de ML
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS features;

-- Tabela de features por célula e hora
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

-- Índices para quality_reports
CREATE INDEX IF NOT EXISTS idx_quality_reports_date ON analytics.quality_reports(report_date);
CREATE INDEX IF NOT EXISTS idx_quality_reports_metrics ON analytics.quality_reports USING GIN(metrics);

-- Tabela de logs de execução do pipeline
CREATE TABLE IF NOT EXISTS analytics.pipeline_logs (
    id SERIAL PRIMARY KEY,
    execution_id UUID UNIQUE NOT NULL,
    started_at TIMESTAMP NOT NULL,
    finished_at TIMESTAMP,
    status VARCHAR(20), -- running, success, failed
    phase VARCHAR(50), -- migrate, spatial_grid, assign_cells, etc
    records_processed INTEGER,
    error_message TEXT,
    execution_time_seconds INTEGER,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Índices para pipeline_logs
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

-- Função para calcular distância entre duas geografias
CREATE OR REPLACE FUNCTION curated.distance_meters(geog1 GEOGRAPHY, geog2 GEOGRAPHY)
RETURNS FLOAT AS $$
BEGIN
    RETURN ST_Distance(geog1, geog2);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Função para verificar se um ponto está dentro do bbox de Campinas
CREATE OR REPLACE FUNCTION curated.is_within_campinas(lat FLOAT, lon FLOAT)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN lat >= -23.1 AND lat <= -22.7 AND lon >= -47.3 AND lon <= -46.8;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Função trigger para atualizar updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Aplicar trigger em tabelas relevantes
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
-- COMENTÁRIOS PARA DOCUMENTAÇÃO
-- ============================================================================

COMMENT ON SCHEMA curated IS 'Dados processados e curados de incidentes criminais';
COMMENT ON SCHEMA external IS 'Dados externos que influenciam a criminalidade';
COMMENT ON SCHEMA features IS 'Features engenheiradas para modelos de ML';
COMMENT ON SCHEMA analytics IS 'Metadados e métricas de qualidade';

COMMENT ON TABLE curated.incidents IS 'Incidentes criminais migrados do banco legado';
COMMENT ON TABLE curated.cells IS 'Grade espacial de células para agregação geográfica';
COMMENT ON TABLE external.weather IS 'Dados meteorológicos históricos e em tempo real';
COMMENT ON TABLE external.holidays IS 'Calendário de feriados e datas especiais';
COMMENT ON TABLE external.events IS 'Eventos que podem impactar a criminalidade';
COMMENT ON TABLE features.cell_hourly IS 'Features por célula e hora para treinamento de ML';
COMMENT ON TABLE analytics.quality_reports IS 'Relatórios de qualidade da base de conhecimento';
COMMENT ON TABLE analytics.pipeline_logs IS 'Logs de execução do pipeline de geração';

-- ============================================================================
-- GRANTS (ajustar conforme suas necessidades de segurança)
-- ============================================================================

-- Conceder acesso ao usuário da aplicação (substitua 'app_user' pelo seu usuário)
-- GRANT USAGE ON SCHEMA curated, external, features, analytics TO app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA curated TO app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA external TO app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA features TO app_user;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA analytics TO app_user;
-- GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA external, features, analytics TO app_user;

-- ============================================================================
-- FIM DAS MIGRATIONS
-- ============================================================================

-- Inserir registro de migração bem-sucedida
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = 'schema_migrations') THEN
        CREATE TABLE public.schema_migrations (
            version VARCHAR(50) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT NOW()
        );
    END IF;
    
    INSERT INTO public.schema_migrations (version, applied_at)
    VALUES ('knowledge_base_v1.0.0', NOW())
    ON CONFLICT (version) DO NOTHING;
END $$;

-- Log de sucesso
DO $$
BEGIN
    RAISE NOTICE '✅ Knowledge Base Schema Migrations aplicadas com sucesso!';
    RAISE NOTICE '📊 Schemas criados: curated, external, features, analytics';
    RAISE NOTICE '🗂️  Tabelas criadas: 9 tabelas principais + views + functions';
    RAISE NOTICE '🚀 Sistema pronto para geração da base de conhecimento!';
END $$;
