import pandas as pd
import psycopg2

def get_connection():
    return psycopg2.connect(
        host="db",
        database="radar",
        user="postgres",
        password="postgres"
    ) # ajustar as credenciais posteriomente

def get_base_knowledge():
    conn = get_connection()
    query = "SELECT * FROM ;" # precisamos substituir pelo nome da tabela
    df = pd.read_sql(query, conn)
    conn.close()
    return df

def save_predictions(df):
    conn = get_connection()
    cursor = conn.cursor()

    cursor.execute("DELETE FROM predictCrime ;") # para limpar os dados antigos
    for _, row in df.iterrows():
        cursor.execute(
            "INSERT INTO predictCrime (neighborhood, risk_level) VALUES (%s, %s)",
            (row['neighborhood'], int(row['risk_level']))
        )
    conn.commit()
    conn.close()
