#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Script para enviar ocorrências criminais via POST para a API
Envia cada JSON individualmente para http://localhost:8080/api/v1/reports/process-text
"""

import json
import requests
import time
from pathlib import Path

# Configurações
API_URL = "http://localhost:8000/api/v1/reports/process-text"
JSON_FILE = "output_json/all_crimes_consolidated.json"
DELAY_BETWEEN_REQUESTS = 0.001  # segundos entre cada requisição (evita sobrecarregar)

def send_crime_report(crime_data: dict, index: int) -> bool:
    """
    Envia um crime individual para a API
    Retorna True se sucesso, False se erro
    """
    try:
        response = requests.post(
            API_URL,
            json=crime_data,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        if response.status_code in [200, 201]:
            print(f"✅ [{index}] Enviado: {crime_data['crime_name']} - {crime_data['report_date']}")
            return True
        else:
            print(f"❌ [{index}] ERRO {response.status_code}: {response.text[:100]}")
            return False
            
    except requests.exceptions.ConnectionError:
        print(f"❌ [{index}] ERRO: Não foi possível conectar à API em {API_URL}")
        return False
    except requests.exceptions.Timeout:
        print(f"⏱️ [{index}] TIMEOUT: Requisição demorou mais de 10s")
        return False
    except Exception as e:
        print(f"❌ [{index}] ERRO: {str(e)}")
        return False


def main():
    print("\n" + "="*80)
    print("ENVIO DE OCORRÊNCIAS CRIMINAIS PARA API")
    print("="*80)
    print(f"🎯 API: {API_URL}")
    print(f"📂 Arquivo: {JSON_FILE}\n")
    
    # Verifica se o arquivo existe
    if not Path(JSON_FILE).exists():
        print(f"❌ ERRO: Arquivo não encontrado: {JSON_FILE}")
        return
    
    # Carrega os crimes
    with open(JSON_FILE, 'r', encoding='utf-8') as f:
        crimes = json.load(f)
    
    total = len(crimes)
    print(f"📊 Total de ocorrências a enviar: {total}\n")
    
    # Confirmação
    resposta = input("Deseja continuar? (s/n): ").strip().lower()
    if resposta != 's':
        print("❌ Operação cancelada pelo usuário")
        return
    
    print("\n🚀 Iniciando envio...\n")
    
    # Contadores
    success_count = 0
    error_count = 0
    start_time = time.time()
    
    # Envia cada crime
    for idx, crime in enumerate(crimes, 1):
        success = send_crime_report(crime, idx)
        
        if success:
            success_count += 1
        else:
            error_count += 1
        
        # Delay entre requisições
        if idx < total:
            time.sleep(DELAY_BETWEEN_REQUESTS)
        
        # Mostra progresso a cada 100 registros
        if idx % 100 == 0:
            elapsed = time.time() - start_time
            rate = idx / elapsed
            remaining = (total - idx) / rate if rate > 0 else 0
            print(f"\n📈 Progresso: {idx}/{total} ({idx/total*100:.1f}%) - "
                  f"Taxa: {rate:.1f} req/s - Tempo restante: {remaining:.0f}s\n")
    
    # Resumo final
    elapsed_total = time.time() - start_time
    print("\n" + "="*80)
    print("✅ ENVIO CONCLUÍDO")
    print("="*80)
    print(f"📊 Total enviado: {total}")
    print(f"✅ Sucessos: {success_count}")
    print(f"❌ Erros: {error_count}")
    print(f"⏱️  Tempo total: {elapsed_total:.2f}s")
    print(f"📈 Taxa média: {total/elapsed_total:.2f} requisições/segundo")
    print("="*80 + "\n")


if __name__ == "__main__":
    main()