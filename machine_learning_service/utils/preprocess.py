import pandas as pd
import numpy as np

def prepare_data(df):
    df = df.dropna()
    
    feature_cols = [
        'lag_1h', 'lag_24h', 'lag_7d',
        'dow', 'hour',
        'holiday', 'is_weekend', 'is_business_hours',
        'center_lat', 'center_lng'
    ]
    
    X = df[feature_cols].copy()
    
    # Classificar y_count em níveis de risco
    y = pd.cut(
        df['y_count'],
        bins=[-1, 0, 2, 5, np.inf],
        labels=[0, 1, 2, 3]
    ).astype(int)
    
    cell_ids = df['cell_id']
    
    return X, y, cell_ids