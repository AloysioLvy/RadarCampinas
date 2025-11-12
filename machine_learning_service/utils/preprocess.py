import pandas as pd

def prepare_data(df):
    #  selecionar features e target
    X = df[['category', 'severity', 'lon', 'lat', 'confidence']].copy()

    # categorias em números
    X['category'] = X['category'].astype('category').cat.codes

    #  target como inteiro
    y = df['severity'].astype(int)

    neighborhoods = df['neighborhood']

    return X, y, neighborhoods
