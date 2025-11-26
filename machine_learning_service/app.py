from fastapi import FastAPI, Query
from utils.db import get_base_knowledge, get_monthly_predictions_with_coords, save_predictions, train_model_monthly
from utils.preprocess import prepare_data
from sklearn.ensemble import RandomForestClassifier
import pandas as pd
import logging
from fastapi.middleware.cors import CORSMiddleware

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")

app = FastAPI()

# ---------------------------------------------------------
# 🔥 Habilitando CORS
# ---------------------------------------------------------
origins = [
    "http://localhost:3000",
    "http://127.0.0.1:3000",
    "https://tkrkwptv-8000.brs.devtunnels.ms"
]

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)
# ---------------------------------------------------------

@app.post("/training")
def train_model():
    logging.info("Iniciando treinamento do modelo...")

    try:
        df = get_base_knowledge()
    except Exception as e:
        logging.error(f"Erro ao ler a base de conhecimento: {e}")
        return {"error": "Falha ao acessar a base de conhecimento"}

    if df.empty:
        logging.warning("Base de conhecimento vazia")
        return {"error": "Base de conhecimento vazia"}

    X, y, neighborhoods = prepare_data(df)
    logging.info(f"{len(df)} linhas carregadas para treinamento")

    model = RandomForestClassifier(n_estimators=100, random_state=42)
    model.fit(X, y)
    logging.info("Modelo treinado com sucesso")

    y_pred = model.predict(X)

    predict_df = pd.DataFrame({
        "neighborhood": neighborhoods,
        "risk_level": y_pred
    })

    try:
        predict_agg = (
            predict_df
            .groupby("neighborhood", as_index=False)
            .agg({"risk_level": "max"})
        )
        save_predictions(predict_agg)
        logging.info(f"{len(predict_df)} previsões salvas no banco")
    except Exception as e:
        logging.error(f"Erro ao salvar previsões: {e}")
        return {"error": "Falha ao salvar previsões no banco"}

    return {"message": "Treinamento concluído", "rows": len(predict_df)}


# ---------------------------------------------------------
# 🔥 POST /training-monthly → SÓ TREINA E SALVA
# ---------------------------------------------------------
@app.post("/training-monthly")
def training_monthly_endpoint(
    year: int = Query(..., ge=2000, le=2100),
    month: int = Query(..., ge=1, le=12)
):
    """
    Treina o modelo mensal e salva as previsões no banco.
    Não retorna as previsões (use GET /predictions para isso).
    """
    train_result = train_model_monthly(target_year=year, target_month=month)

    if isinstance(train_result, dict) and "error" in train_result:
        return train_result

    return {
        "year": year,
        "month": month,
        "summary": train_result
    }


# ---------------------------------------------------------
# 🔥 GET /predictions → SÓ BUSCA E RETORNA
# ---------------------------------------------------------
@app.get("/predictions")
def get_predictions_endpoint(
    year: int = Query(..., ge=2000, le=2100),
    month: int = Query(..., ge=1, le=12)
):
    """
    Retorna as previsões mensais já salvas no banco.
    Formato: [{neighborhood, lat, long, risk_level}, ...]
    """
    predictions = get_monthly_predictions_with_coords(year, month)

    if not predictions:
        return {
            "year": year,
            "month": month,
            "predictions": [],
            "message": "Nenhuma previsão encontrada para este mês. Execute POST /training-monthly primeiro."
        }

    return {
        "year": year,
        "month": month,
        "predictions": predictions
    }