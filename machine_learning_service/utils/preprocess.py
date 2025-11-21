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

    return X, y, neighborhoods