import datetime 
import pandas as pd
import logging
import pyodbc
from sklearn.ensemble import RandomForestRegressor
import numpy as np
from .preprocess import prepare_data_monthly
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

def get_base_knowledge_monthly(months_back=24):
    """
    Carrega features mensais (célula-mês) para treinamento.
    
    Args:
        months_back: quantos meses para trás buscar (padrão: 24 meses = 2 anos)
    
    Returns:
        DataFrame com features mensais por célula
    """
    conn = get_connection()
    try:
        # Calcular ano-mês de corte
        now = datetime.datetime.now()
        cutoff_year_month = (now.year * 100 + now.month) - months_back
        
        query = f"""
            SELECT 
                fcm.cell_id,
                cn.neighborhood,
                fcm.year,
                fcm.month,
                fcm.y_count_month,
                fcm.lag_1m,
                fcm.lag_3m,
                cc.center_lat,
                cc.center_lng
            FROM features_cell_monthly fcm
            INNER JOIN curated_cells cc 
                ON fcm.cell_id = cc.cell_id
            LEFT JOIN cell_neighborhoods cn
                ON cn.cell_id = fcm.cell_id
            WHERE (fcm.year * 100 + fcm.month) >= {cutoff_year_month}
            ORDER BY fcm.year, fcm.month, fcm.cell_id
        """
        
        df = pd.read_sql(query, conn)
        logging.info(f"✅ Carregados {len(df)} registros mensais da base de conhecimento")
        return df
        
    finally:
        conn.close()
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
def save_predictions(predictions_df: pd.DataFrame):
    """
    Salva previsões na tabela predict_crimes.
    
    Args:
        predictions_df: DataFrame com colunas:
            - neighborhood
            - risk_level
            - prediction_date
    """
    logging.info("🔄 Iniciando salvamento de previsões...")
    logging.info(f"📊 Total de previsões a salvar: {len(predictions_df)}")

    conn = get_connection()
    cursor = conn.cursor()

    try:
    

        # Preparar dados
        logging.info("📝 Preparando dados para inserção...")
        records = []
        for _, row in predictions_df.iterrows():
            records.append((
                row['neighborhood'],
                int(row['risk_level']),
                row['prediction_date'],
                'monthly'  # model_type
            ))
        logging.info(f"✅ {len(records)} registros preparados")

        # Inserir
        logging.info("💾 Inserindo novas previsões no banco...")
        insert_query = """
        INSERT INTO predict_crimes (neighborhood, risk_level, prediction_date, model_type)
        VALUES (?, ?, ?, ?)
        """
        cursor.executemany(insert_query, records)

        logging.info("💿 Confirmando transação...")
        conn.commit()
        logging.info(f"✅ {len(records)} previsões salvas com sucesso em predict_crimes!")

    except Exception as e:
        logging.error(f"❌ Erro ao salvar previsões: {e}")
        conn.rollback()
        raise
    finally:
        logging.info("🔒 Fechando cursor e conexão...")
        cursor.close()
        conn.close()
        logging.info("✅ Recursos liberados")
def get_features_for_next_month(target_month: datetime) -> pd.DataFrame:
    """
    Retorna features para o próximo mês (M+1), baseado nos lags.
    
    Args:
        target_month: mês alvo (ex: 2024-12-01)
    
    Returns:
        DataFrame com features para cada célula no mês alvo
    """
    year = target_month.year
    month = target_month.month
    
    query = f"""
    SELECT
        cell_id,
        {year} AS year,
        {month} AS month,
        
        -- lag_1m: crimes no mês anterior (M-1)
        ISNULL((
            SELECT y_count_month
            FROM features_cell_monthly f0
            WHERE f0.cell_id = f.cell_id
              AND f0.year = {year if month > 1 else year - 1}
              AND f0.month = {month - 1 if month > 1 else 12}
        ), 0) AS lag_1m,
        
        -- lag_3m: soma dos últimos 3 meses (M-3 até M-1)
        ISNULL((
            SELECT SUM(y_count_month)
            FROM features_cell_monthly f0
            WHERE f0.cell_id = f.cell_id
              AND DATEFROMPARTS(f0.year, f0.month, 1)
                >= DATEADD(month, -3, DATEFROMPARTS({year}, {month}, 1))
              AND DATEFROMPARTS(f0.year, f0.month, 1)
                < DATEFROMPARTS({year}, {month}, 1)
        ), 0) AS lag_3m
        
    FROM (
        SELECT DISTINCT cell_id
        FROM features_cell_monthly
    ) f
    """
    
    conn = get_connection()
    df = pd.read_sql(query, conn)
    conn.close()
    
    # Adicionar bairro (join com cell_neighborhoods)
    conn = get_connection()
    neighborhoods_df = pd.read_sql("SELECT cell_id, neighborhood FROM cell_neighborhoods", conn)
    conn.close()
    
    df = df.merge(neighborhoods_df, on='cell_id', how='left')
    
    return df

def train_model_monthly(target_year: int, target_month: int):
    """
    Treina modelo mensal e gera previsões para um mês específico (target_year, target_month).
    Ex: (2024, 12) → previsões para dezembro/2024.
    """
    logging.info(f"🗓️  Iniciando treinamento do modelo MENSAL para {target_year}-{target_month:02d}...")

    # 1) Ler base de conhecimento (histórico) - últimos 24 meses, por exemplo
    try:
        df = get_base_knowledge_monthly(months_back=24)
    except Exception as e:
        logging.error(f"Erro ao ler a base de conhecimento mensal: {e}")
        return {"error": "Falha ao acessar a base de conhecimento mensal"}

    if df.empty:
        logging.warning("Base de conhecimento mensal vazia")
        return {"error": "Base de conhecimento mensal vazia"}

    # Log da distribuição do target
    logging.info("Distribuição de y_count_month:")
    logging.info(df['y_count_month'].value_counts().head(20))
    logging.info(f"Min: {df['y_count_month'].min()}, Max: {df['y_count_month'].max()}")

    # 2) Preparar features e target para treino
    X, y, neighborhoods = prepare_data_monthly(df)
    logging.info(f"{len(df)} linhas carregadas para treinamento mensal")

    # 3) Treinar modelo de regressão
    model = RandomForestRegressor(
        n_estimators=200,
        random_state=42,
        max_depth=10,
        min_samples_split=5
    )
    model.fit(X, y)
    logging.info("✅ Modelo mensal treinado com sucesso")

    # Previsões (contagem de crimes) no conjunto de TREINO
    y_pred = model.predict(X)

    # ========================================
    # AVALIAÇÃO DO MODELO
    # ========================================
    from sklearn.metrics import mean_absolute_error, mean_squared_error, r2_score

    mae = mean_absolute_error(y, y_pred)
    mse = mean_squared_error(y, y_pred)
    rmse = np.sqrt(mse) # np já importado
    r2 = r2_score(y, y_pred)

    logging.info(f"📊 Métricas de Avaliação (no conjunto de treino):")
    logging.info(f"  MAE (Erro Médio Absoluto): {mae:.2f}")
    logging.info(f"  RMSE (Raiz do Erro Quadrático Médio): {rmse:.2f}")
    logging.info(f"  R² (Coeficiente de Determinação): {r2:.2f}")

    # ========================================
    # VALIDAÇÃO CRUZADA
    # ========================================
    from sklearn.model_selection import cross_val_score
    from sklearn.metrics import make_scorer

    # Definir scorers para as métricas que você quer
    mae_scorer = make_scorer(mean_absolute_error, greater_is_better=False) # MAE é melhor qdo menor
    mse_scorer = make_scorer(mean_squared_error, greater_is_better=False)
    r2_scorer = make_scorer(r2_score)

    # Rodar validação cruzada (ex: 5 folds)
    cv_mae_scores = cross_val_score(model, X, y, cv=5, scoring=mae_scorer)
    cv_rmse_scores = np.sqrt(-cross_val_score(model, X, y, cv=5, scoring=mse_scorer)) # MSE é negativo pq greater_is_better=False
    cv_r2_scores = cross_val_score(model, X, y, cv=5, scoring=r2_scorer)

    logging.info(f"📊 Métricas de Avaliação (Validação Cruzada - 5 folds):")
    logging.info(f"  MAE Médio: {-cv_mae_scores.mean():.2f} (+/- {-cv_mae_scores.std():.2f})")
    logging.info(f"  RMSE Médio: {cv_rmse_scores.mean():.2f} (+/- {cv_rmse_scores.std():.2f})")
    logging.info(f"  R² Médio: {cv_r2_scores.mean():.2f} (+/- {cv_r2_scores.std():.2f})")

    # Treinar o modelo final com todos os dados de treino
    model.fit(X, y)
    logging.info("✅ Modelo mensal treinado com sucesso (após validação cruzada)")

    # 4) Gerar features para o mês alvo
    try:
        target_month_date = datetime.datetime(target_year, target_month, 1)
    except ValueError as e:
        logging.error(f"Mês/ano inválido: {e}")
        return {"error": "Mês/ano inválido"}

    logging.info(f"📅 Gerando features para o mês alvo: {target_month_date.strftime('%Y-%m-%d')}")

    try:
        df_target = get_features_for_month(target_month_date)
    except Exception as e:
        logging.error(f"Erro ao buscar features para o mês alvo: {e}")
        return {"error": "Falha ao buscar features para o mês alvo"}

    if df_target.empty:
        logging.warning("Não há features para o mês alvo (get_features_for_month retornou vazio)")
        return {"error": "Features para o mês alvo não disponíveis"}

    feature_cols = ['lag_1m', 'lag_3m', 'month']

    X_target = df_target[feature_cols]

    # usar bairro se disponível, senão cell_id
    neighborhoods_target = df_target['neighborhood'].fillna(df_target['cell_id'])

    # 5) Prever contagem de crimes para o mês alvo
    y_pred = model.predict(X_target)

    # 6) Transformar em níveis de risco usando percentis dos próprios y_pred
    p60 = np.percentile(y_pred, 60)
    p85 = np.percentile(y_pred, 85)

    def to_risk_level(v):
        if v <= p60:
            return 0  # baixo
        elif v <= p85:
            return 1  # médio
        else:
            return 2  # alto

    risk_levels = [to_risk_level(v) for v in y_pred]

    logging.info(f"Thresholds de risco: baixo <= {p60:.2f}, médio <= {p85:.2f}, alto > {p85:.2f}")

    # 7) Montar DataFrame de previsões
    predict_df = pd.DataFrame({
        "neighborhood": neighborhoods_target,
        "risk_level": risk_levels,
        "predicted_count": y_pred,
        "prediction_date": target_month_date.strftime('%Y-%m-%d')
    })

    logging.info("Distribuição de risk_level:")
    logging.info(predict_df['risk_level'].value_counts())

    # 8) Agregar por bairro (opcional, se você quiser 1 linha por bairro)
    predict_agg = (
        predict_df
        .groupby("neighborhood", as_index=False)
        .agg({
            "risk_level": "max",
            "predicted_count": "sum",
            "prediction_date": "first"
        })
    ) 
    # ------------------------------------------------------------------
    # 8.5) AJUSTE DE RISCO PARA MÊS SEM DADOS (todas previsões = 0)
    # ------------------------------------------------------------------
    if predict_agg['predicted_count'].max() == 0:
        logging.info("Todas as previsões do mês alvo são 0. Usando histórico de 12 meses para definir risco relativo.")

        # Garantir que df (base histórica) tenha uma coluna de data mensal
        df_hist = df.copy()

        if 'date' not in df_hist.columns:
            # Assume que df tem colunas 'year' e 'month'
            df_hist['date'] = pd.to_datetime(
                dict(year=df_hist['year'], month=df_hist['month'], day=1)
            )

        # Janela de 12 meses antes do mês alvo
        start_date = target_month_date - pd.DateOffset(months=12)
        end_date = target_month_date  # exclusivo

        mask = (df_hist['date'] >= start_date) & (df_hist['date'] < end_date)

        # Somar crimes por bairro nos últimos 12 meses
        # df_hist precisa ter 'neighborhood' (ou usamos cell_id e depois juntamos)
        hist_agg = (
            df_hist[mask]
            .groupby('neighborhood', as_index=False)['y_count_month']
            .sum()
            .rename(columns={'y_count_month': 'hist_12m'})
        )

        logging.info(f"Histórico 12m calculado para {len(hist_agg)} bairros.")
        logging.info(hist_agg.sort_values('hist_12m', ascending=False).head(10))

        # Juntar histórico em predict_agg
        predict_agg = predict_agg.merge(hist_agg, on='neighborhood', how='left')
        predict_agg['hist_12m'] = predict_agg['hist_12m'].fillna(0)

        # Ordenar por histórico de crimes (do menor para o maior)
        predict_agg = predict_agg.sort_values('hist_12m').reset_index(drop=True)

        n = len(predict_agg)
        p60_idx = int(n * 0.6)
        p85_idx = int(n * 0.85)

        # Redefinir risk_level com base no histórico
        predict_agg['risk_level'] = 0
        if n > 0:
            # médio: 60% até 85%
            predict_agg.loc[p60_idx:p85_idx-1, 'risk_level'] = 1
            # alto: top 15%
            predict_agg.loc[p85_idx:, 'risk_level'] = 2

        logging.info("Distribuição de risk_level (ajustada por histórico 12m):")
        logging.info(predict_agg['risk_level'].value_counts())


    # 9) Salvar previsões no banco
    try:
        save_predictions(predict_agg)
        logging.info(f"✅ {len(predict_agg)} previsões mensais salvas no banco")
    except Exception as e:
        logging.error(f"Erro ao salvar previsões: {e}")
        return {"error": "Falha ao salvar previsões no banco"}

    return {
        "message": "Treinamento mensal concluído",
        "target_year": target_year,
        "target_month": target_month,
        "rows_trained": len(df),
        "rows_predicted": len(predict_df),
        "neighborhoods": len(predict_agg),
        "risk_distribution": predict_agg['risk_level'].value_counts().to_dict()
    }

def get_features_for_month(target_month: datetime) -> pd.DataFrame:
    """
    Retorna features (lag_1m, lag_3m, month, neighborhood/cell_id) para o mês alvo.
    """
    year = target_month.year
    month = target_month.month

    query = f"""
    WITH Cells AS (
        SELECT DISTINCT cell_id
        FROM features_cell_monthly
    )
    SELECT
        c.cell_id,
        {year} AS year,
        {month} AS month,
        -- lag_1m: crimes no mês anterior
        ISNULL((
            SELECT y_count_month
            FROM features_cell_monthly f1
            WHERE f1.cell_id = c.cell_id
              AND f1.year = {year if month > 1 else year - 1}
              AND f1.month = {month - 1 if month > 1 else 12}
        ), 0) AS lag_1m,
        -- lag_3m: soma últimos 3 meses
        ISNULL((
            SELECT SUM(y_count_month)
            FROM features_cell_monthly f3
            WHERE f3.cell_id = c.cell_id
              AND DATEFROMPARTS(f3.year, f3.month, 1)
                >= DATEADD(month, -3, DATEFROMPARTS({year}, {month}, 1))
              AND DATEFROMPARTS(f3.year, f3.month, 1)
                <  DATEFROMPARTS({year}, {month}, 1)
        ), 0) AS lag_3m
    FROM Cells c
    """
    conn = get_connection()
    df = pd.read_sql(query, conn)
    conn.close()

    # Adicionar bairro se você tiver uma tabela de mapeamento cell -> neighborhood
    conn = get_connection()
    map_df = pd.read_sql("SELECT cell_id, neighborhood FROM cell_neighborhoods", conn)
    conn.close()

    df = df.merge(map_df, on="cell_id", how="left")

    return df
def get_monthly_predictions_with_coords(target_year: int, target_month: int):
    """
    Retorna previsões mensais já salvas em predict_crimes
    no formato: [{neighborhood, lat, long, risk_level}, ...]
    """

    prediction_date = f"{target_year:04d}-{target_month:02d}-01"

    conn = get_connection()
    cursor = conn.cursor()

    query = """
    SELECT
        pc.neighborhood,
        n.latitude AS lat,
        n.longitude AS long,
        pc.risk_level
    FROM predict_crimes pc
    LEFT JOIN neighborhoods n
        ON pc.neighborhood = n.name
    WHERE pc.prediction_date = ?
      AND pc.model_type = 'monthly'
    ORDER BY pc.risk_level DESC, pc.neighborhood;
    """

    df = pd.read_sql(query, conn, params=[prediction_date])
    conn.close()

    # Converter para lista de dicts no formato desejado
    results = []
    for _, row in df.iterrows():
        results.append({
            "neighborhood": row["neighborhood"],
            "lat": float(row["lat"]) if row["lat"] is not None else None,
            "long": float(row["long"]) if row["long"] is not None else None,
            "risk_level": int(row["risk_level"]),
        })

    return results