import pandas as pd
import mysql.connector
import logging

def get_connection():
    try:
        logging.info("Tentando conectar ao banco MySQL...")
        conn = mysql.connector.connect(
            host="regulus.cotuca.unicamp.br",
            port=3306,
            database="BD24452",
            user="BD24452",
            password="BD24452"
        )
        logging.info("Conexão estabelecida com sucesso!")
        return conn
    except mysql.connector.Error as err:
        logging.error(f"Erro ao conectar ao banco: {err}")
        raise

def get_base_knowledge():
    conn = get_connection()
    try:
        query = "SELECT * FROM autor;"  #trocar pelo nome 
        df = pd.read_sql(query, conn)
        return df
    finally:
        conn.close()

def save_predictions(df):
    conn = get_connection()
    cursor = conn.cursor()
    try:
        # limpa os dados antigos
        cursor.execute("DELETE FROM predictCrime;")
        
        # insere as novas previsões
        insert_query = "INSERT INTO predictCrime (neighborhood, risk_level) VALUES (%s, %s)"
        data = [(row['neighborhood'], int(row['risk_level'])) for _, row in df.iterrows()]
        cursor.executemany(insert_query, data)
        
        conn.commit()
    finally:
        cursor.close()
        conn.close()
