-- ============================================================================
-- KNOWLEDGE BASE SCHEMA MIGRATIONS - MYSQL VERSION
-- TCC Radar Campinas - Base de Conhecimento para IA Preditiva
-- ============================================================================

-- ============================================================================
-- SCHEMA: curated
-- ============================================================================

CREATE DATABASE IF NOT EXISTS BD24452;
USE BD24452;

-- Tabela de incidentes criminais
CREATE TABLE IF NOT EXISTS curated_incidents (
    id VARCHAR(50) PRIMARY KEY,
    occurred_at DATETIME NOT NULL,
    category VARCHAR(50) NOT NULL,
    severity INT NOT NULL CHECK (severity BETWEEN 1 AND 10),
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    neighborhood VARCHAR(100),
    confidence FLOAT CHECK (confidence BETWEEN 0 AND 1),
    source VARCHAR(50) DEFAULT 'legacy_reports',
    cell_id VARCHAR(50),
    cell_resolution INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_occurred_at (occurred_at),
    INDEX idx_cell_id (cell_id),
    INDEX idx_category (category),
    INDEX idx_severity (severity),
    INDEX idx_lat_lng (latitude, longitude)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tabela de células da grade espacial
CREATE TABLE IF NOT EXISTS curated_cells (
    cell_id VARCHAR(50) PRIMARY KEY,
    cell_resolution INT NOT NULL,
    city VARCHAR(50) DEFAULT 'Campinas',
    center_lat DECIMAL(10, 8) NOT NULL,
    center_lng DECIMAL(11, 8) NOT NULL,
    bounds_json JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_resolution (cell_resolution),
    INDEX idx_city (city),
    INDEX idx_center (center_lat, center_lng)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- SCHEMA: external (Feriados)
-- ============================================================================

-- Tabela de feriados
CREATE TABLE IF NOT EXISTS external_holidays (
    id INT AUTO_INCREMENT PRIMARY KEY,
    date DATE NOT NULL,
    name VARCHAR(100) NOT NULL,
    type VARCHAR(50),
    city VARCHAR(50) DEFAULT 'Campinas',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY unique_date_city (date, city),
    INDEX idx_date (date),
    INDEX idx_city (city)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Popular feriados fixos
INSERT IGNORE INTO external_holidays (date, name, type, city) VALUES
    -- 2024
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
    -- 2025
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
    -- 2026
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
    ('2026-12-25', 'Natal', 'nacional', 'Campinas');

-- ============================================================================
-- SCHEMA: features
-- ============================================================================

-- Tabela de features por célula e hora
CREATE TABLE IF NOT EXISTS features_cell_hourly (
    id INT AUTO_INCREMENT PRIMARY KEY,
    cell_id VARCHAR(50) NOT NULL,
    ts DATETIME NOT NULL,
    
    -- Target variable
    y_count INT DEFAULT 0,
    
    -- Lag features
    lag_1h INT DEFAULT 0,
    lag_24h INT DEFAULT 0,
    lag_7d INT DEFAULT 0,
    
    -- Rolling window features
    roll_3h_sum INT DEFAULT 0,
    roll_24h_sum INT DEFAULT 0,
    roll_7d_sum INT DEFAULT 0,
    roll_7d_avg FLOAT,
    roll_7d_std FLOAT,
    
    -- Temporal features
    dow INT, -- day of week (0-6)
    hour INT, -- hour of day (0-23)
    is_weekend BOOLEAN,
    is_business_hours BOOLEAN,
    
    -- Calendar features
    holiday BOOLEAN DEFAULT FALSE,
    day_before_holiday BOOLEAN DEFAULT FALSE,
    day_after_holiday BOOLEAN DEFAULT FALSE,
    
    -- Spatial features
    neighbor_avg_crime FLOAT,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    UNIQUE KEY unique_cell_ts (cell_id, ts),
    INDEX idx_cell_id (cell_id),
    INDEX idx_ts (ts),
    INDEX idx_dow (dow),
    INDEX idx_hour (hour)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- SCHEMA: analytics
-- ============================================================================

-- Tabela de relatórios de qualidade
CREATE TABLE IF NOT EXISTS analytics_quality_reports (
    id INT AUTO_INCREMENT PRIMARY KEY,
    report_date DATE UNIQUE NOT NULL,
    metrics JSON NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_report_date (report_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Tabela de logs de execução do pipeline
CREATE TABLE IF NOT EXISTS analytics_pipeline_logs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    execution_id VARCHAR(36) UNIQUE NOT NULL,
    started_at DATETIME NOT NULL,
    finished_at DATETIME,
    status VARCHAR(20),
    phase VARCHAR(50),
    records_processed INT,
    error_message TEXT,
    execution_time_seconds INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_execution_id (execution_id),
    INDEX idx_started_at (started_at),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- VIEWS
-- ============================================================================

-- View para estatísticas de incidentes por célula
CREATE OR REPLACE VIEW analytics_cell_statistics AS
SELECT 
    cell_id,
    COUNT(*) as total_incidents,
    COUNT(DISTINCT DATE(occurred_at)) as days_with_incidents,
    AVG(severity) as avg_severity,
    MIN(occurred_at) as first_incident,
    MAX(occurred_at) as last_incident
FROM curated_incidents
WHERE cell_id IS NOT NULL
GROUP BY cell_id;

-- View para cobertura temporal
CREATE OR REPLACE VIEW analytics_temporal_coverage AS
SELECT 
    DATE(occurred_at) as date,
    COUNT(*) as incidents,
    COUNT(DISTINCT cell_id) as cells_affected,
    AVG(severity) as avg_severity
FROM curated_incidents
GROUP BY DATE(occurred_at)
ORDER BY date;

-- ============================================================================
-- MIGRATIONS TABLE
-- ============================================================================

CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(50) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT IGNORE INTO schema_migrations (version) VALUES ('knowledge_base_v1.1.0_mysql');

SELECT '✅ Knowledge Base Schema Migrations (MySQL) aplicadas com sucesso!' as status;