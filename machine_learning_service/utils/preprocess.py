import pandas as pd
import numpy as np

def prepare_data(df: pd.DataFrame):
    feature_cols = [
        'lag_1h', 'lag_24h', 'lag_7d',
        'dow', 'hour',
        'holiday', 'is_weekend', 'is_business_hours'
    ]
    X = df[feature_cols]
    y = df['y_count']  # ou outra coluna de alvo
    neighborhoods = df['neighborhood']  # agora é nome de bairro mesmo
    print(df['y_count'].value_counts().head(20))
    print(df['y_count'].min(), df['y_count'].max())

    return X, y, neighborhoods

def prepare_data_monthly(df: pd.DataFrame):
    """
    Prepara features e target para modelo mensal.
    
    Args:
        df: DataFrame com features mensais
    
    Returns:
        X: features
        y: target (contagem de crimes no mês)
        neighborhoods: identificador (bairro ou cell_id)
    """
    feature_cols = [
        'lag_1m',   # crimes no mês anterior
        'lag_3m',   # crimes nos últimos 3 meses
        'month',    # mês do ano (1-12) - sazonalidade
    ]
    
    X = df[feature_cols]
    y = df['y_count_month']  # contagem de crimes no mês
    
    # usar bairro se disponível, senão cell_id
    neighborhoods = df['neighborhood'].fillna(df['cell_id'])
    
    return X, y, neighborhoods


