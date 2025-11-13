from fastapi import FastAPI
from utils.db import get_base_knowledge, save_predictions
from utils.preprocess import prepare_data
from sklearn.ensemble import RandomForestClassifier
import pandas as pd
import logging

# Configuração básica de logging
logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

app = FastAPI()

@app.post("/training")
def train_model():
    logging.info("Iniciando treinamento do modelo...")

    # Ler a base de conhecimento
    try:
        df = get_base_knowledge()
    except Exception as e:
        logging.error(f"Erro ao ler a base de conhecimento: {e}")
        return {"error": "Falha ao acessar a base de conhecimento"}

    if df.empty:
        logging.warning("Base de conhecimento vazia")
        return {"error": "Base de conhecimento vazia"}

    # Preparar features e target
    X, y, neighborhoods = prepare_data(df)
    logging.info(f"{len(df)} linhas carregadas para treinamento")

    # Treinar modelo
    model = RandomForestClassifier(n_estimators=100, random_state=42)
    model.fit(X, y)
    logging.info("Modelo treinado com sucesso")

    # Previsões
    y_pred = model.predict(X)

    # Gerando tabela de previsões
    predict_df = pd.DataFrame({
        "neighborhood": neighborhoods,
        "risk_level": y_pred
    })

    # Salvar previsões no banco
    try:
        save_predictions(predict_df)
        logging.info(f"{len(predict_df)} previsões salvas no banco")
    except Exception as e:
        logging.error(f"Erro ao salvar previsões: {e}")
        return {"error": "Falha ao salvar previsões no banco"}

    return {"message": "Treinamento concluído", "rows": len(predict_df)}
