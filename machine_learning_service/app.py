from fastapi import FastAPI
from utils.db import get_base_knowledge, save_predictions
from utils.preprocess import prepare_data
from sklearn.ensemble import RandomForestClassifier
import pandas as pd

app = FastAPI()

@app.post("/training")
def train_model():
    # le a kb
    df = get_base_knowledge()

    if df.empty:
        return {"error": "Base de conhecimento vazia"}

    # preparar features e target
    X, y, neighborhoods = prepare_data(df)

    #  treinar modelo de fato
    model = RandomForestClassifier(n_estimators=100, random_state=42)
    model.fit(X, y)

    # previsoes
    y_pred = model.predict(X)

    #  gerando tabela de previsoes, consumida pelo frontend
    predict_df = pd.DataFrame({
        "neighborhood": neighborhoods,
        "risk_level": y_pred
    })

    # salvando no banco
    save_predictions(predict_df)

    return {"message": "Treinamento concluído", "rows": len(predict_df)}
