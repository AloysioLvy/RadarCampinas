import datetime 
import pandas as pd
import logging
import pyodbc
def get_connection():
    """
    Estabelece conexão com SQL Server usando pyodbc.
    """
    try:
        logging.info("Tentando conectar ao banco SQL Server...")
        
        # String de conexão para SQL Server
        conn_str = (
            "DRIVER={ODBC Driver 17 for SQL Server};"
            "SERVER=regulus.cotuca.unicamp.br,1433;"  # host,porta
            "DATABASE=BD24452;"
            "UID=BD24452;"
            "PWD=BD24452;"
            "TrustServerCertificate=yes;"  # Para evitar erro de certificado SSL
        )
        
        conn = pyodbc.connect(conn_str)
        logging.info("Conexão estabelecida com sucesso!")
        return conn
        
    except pyodbc.Error as err:
        logging.error(f"Erro ao conectar ao banco: {err}")
        raise

def get_base_knowledge(days_back=90):
    conn = get_connection()
    try:
        cutoff_date = (datetime.datetime.now() - datetime.timedelta(days=days_back)).strftime('%Y-%m-%d')
        
        query = f"""
            SELECT 
                fch.cell_id,
                cn.neighborhood,  -- bairro associado à célula
                CAST(fch.ts AS DATETIME) as ts,
                fch.y_count,
                fch.lag_1h,
                fch.lag_24h,
                fch.lag_7d,
                fch.dow,
                fch.hour,
                CAST(fch.holiday AS INT) as holiday,
                CAST(fch.is_weekend AS INT) as is_weekend,
                CAST(fch.is_business_hours AS INT) as is_business_hours,
                cc.center_lat,
                cc.center_lng
            FROM features_cell_hourly fch
            INNER JOIN curated_cells cc 
                ON fch.cell_id = cc.cell_id
            LEFT JOIN cell_neighborhoods cn
                ON cn.cell_id = fch.cell_id
            WHERE fch.ts >= '{cutoff_date}'
            ORDER BY fch.ts, fch.cell_id
        """
        
        df = pd.read_sql(query, conn)
        logging.info(f"✅ Carregados {len(df)} registros da base de conhecimento")
        return df
        
    finally:
        conn.close()

def save_predictions(df):
    logging.info("🔄 Iniciando salvamento de previsões...")
    logging.info(f"📊 Total de previsões a salvar: {len(df)}")
    
    conn = get_connection()
    cursor = conn.cursor()
    
    try:
        # Limpa os dados antigos
        logging.info("🗑️  Removendo previsões antigas da tabela predict_crimes...")
        cursor.execute("DELETE FROM predict_crimes;")
        deleted_count = cursor.rowcount
        logging.info(f"✅ {deleted_count} registros antigos removidos")
        
        # Prepara os dados para inserção
        logging.info("📝 Preparando dados para inserção...")
        insert_query = "INSERT INTO predict_crimes (neighborhood, risk_level) VALUES (?, ?)"
        data = [(row['neighborhood'], int(row['risk_level'])) for _, row in df.iterrows()]
        logging.info(f"✅ {len(data)} registros preparados")
        
        # Insere as novas previsões
        logging.info("💾 Inserindo novas previsões no banco...")
        cursor.fast_executemany = True  # Acelera inserção em lote
        cursor.executemany(insert_query, data)
        
        # Commit das mudanças
        logging.info("💿 Confirmando transação...")
        conn.commit()
        logging.info(f"✅ {len(data)} previsões salvas com sucesso em predict_crimes!")
        
    except Exception as e:
        logging.error(f"❌ Erro ao salvar previsões: {e}")
        logging.info("🔄 Revertendo transação...")
        conn.rollback()
        logging.error("❌ Transação revertida")
        raise
        
    finally:
        logging.info("🔒 Fechando cursor e conexão...")
        cursor.close()
        conn.close()
        logging.info("✅ Recursos liberados")
