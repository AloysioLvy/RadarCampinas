#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Script para processar dados de ocorrências criminais das delegacias de Campinas
Gera JSONs com informações de crimes, coordenadas e pesos
VERSÃO CORRIGIDA - KeyError resolvido
"""

import pandas as pd
import json
import os
import re
from datetime import datetime
from pathlib import Path
import random

# Coordenadas das delegacias de Campinas
DELEGACIAS_COORDS = {
    "01 DP": {"bairro": "Centro", "latitude": -22.9056, "longitude": -47.0608},
    "02 DP": {"bairro": "Vila São Bernardo", "latitude": -22.9234, "longitude": -47.0445},
    "03 DP": {"bairro": "Botafogo", "latitude": -22.8923, "longitude": -47.0712},
    "04 DP": {"bairro": "Vila Nogueira", "latitude": -22.8867, "longitude": -47.0823},
    "05 DP": {"bairro": "Vila Santana", "latitude": -22.9191, "longitude": -47.0712},
    "06 DP": {"bairro": "Jardim Novo Campos Eliseos", "latitude": -22.8856, "longitude": -47.0589},
    "07 DP": {"bairro": "Cidade Universitária", "latitude": -22.8195, "longitude": -47.0658},
    "08 DP": {"bairro": "Conjunto Habitacional Padre Anchieta", "latitude": -22.9458, "longitude": -47.1089},
    "09 DP": {"bairro": "Vila Aeroporto / DIC", "latitude": -22.9389, "longitude": -47.0956},
    "10 DP": {"bairro": "Jardim Primavera", "latitude": -22.8978, "longitude": -47.0345},
    "11 DP": {"bairro": "Jardim Ipaussurama", "latitude": -22.8645, "longitude": -47.1123},
    "12 DP": {"bairro": "Sousas", "latitude": -22.8856, "longitude": -46.9567},
    "13 DP": {"bairro": "Cambuí", "latitude": -22.8989, "longitude": -47.0523},
}

# Crimes hediondos (peso 9)
HEINOUS_CRIMES = [
    "latrocínio",
    "homicídio qualificado",
    "homicídio praticado por grupo de extermínio",
    "feminicídio",
    "genocídio",
    "estupro",
    "estupro de vulnerável",
    "atentado violento ao pudor",
    "favorecimento à prostituição",
    "exploração sexual",
    "tráfico de pessoas",
    "tráfico de drogas",
    "organização criminosa",
    "comércio ilegal de armas",
    "extorsão qualificada",
    "sequestro e cárcere privado",
    "extorsão mediante sequestro",
    "envenenamento de alimentos",
    "epidemia com resultado morte",
    "falsificação de medicamentos",
    "tráfico internacional de armas",
    "sequestro e extorsão qualificada",
    "Homicídio Doloso"
]

# Mapeamento de tipos de crime para nomes padronizados
CRIME_MAPPING = {
    "HOMICÍDIO DOLOSO": "Homicídio Doloso",
    "HOMICÍDIO DOLOSO POR ACIDENTE DE TRÂNSITO": "Homicídio Doloso por Acidente de Trânsito",
    "HOMICÍDIO CULPOSO POR ACIDENTE DE TRÂNSITO": "Homicídio Culposo por Acidente de Trânsito",
    "HOMICÍDIO CULPOSO OUTROS": "Homicídio Culposo",
    "TENTATIVA DE HOMICÍDIO": "Tentativa de Homicídio",
    "LESÃO CORPORAL SEGUIDA DE MORTE": "Lesão Corporal Seguida de Morte",
    "LESÃO CORPORAL DOLOSA": "Lesão Corporal Dolosa",
    "LESÃO CORPORAL CULPOSA POR ACIDENTE DE TRÂNSITO": "Lesão Corporal Culposa por Acidente de Trânsito",
    "LESÃO CORPORAL CULPOSA - OUTRAS": "Lesão Corporal Culposa",
    "LATROCÍNIO": "Latrocínio",
    "TOTAL DE ESTUPRO": "Estupro",
    "ESTUPRO": "Estupro",
    "ESTUPRO DE VULNERÁVEL": "Estupro de Vulnerável",
    "TOTAL DE ROUBO - OUTROS": "Roubo",
    "ROUBO - OUTROS": "Roubo",
    "ROUBO DE VEÍCULO": "Roubo de Veículo",
    "ROUBO A BANCO": "Roubo a Banco",
    "ROUBO DE CARGA": "Roubo de Carga",
    "FURTO - OUTROS": "Furto",
    "FURTO DE VEÍCULO": "Furto de Veículo",
}

# Meses em português
MESES = {
    "Janeiro": 1, "Fevereiro": 2, "Marco": 3, "Março": 3, "Abril": 4,
    "Maio": 5, "Junho": 6, "Julho": 7, "Agosto": 8,
    "Setembro": 9, "Outubro": 10, "Novembro": 11, "Dezembro": 12
}


def calculate_weight_crime(crime_type: str) -> int:
    """
    Calcula o peso do crime baseado na lista de crimes hediondos
    Crimes hediondos: peso 9
    Outros crimes: peso 3
    """
    crime_normalized = crime_type.strip().lower()
    
    for heinous_crime in HEINOUS_CRIMES:
        if heinous_crime.lower() in crime_normalized:
            return 9
    
    return 3


def extract_dp_from_filename(filename: str) -> str:
    """
    Extrai o número da DP do nome do arquivo
    Exemplo: "OcorrenciaMensal(Criminal)-01 DP - Campinas_20251125_223822.xlsx" -> "01 DP"
    """
    match = re.search(r'(\d{2})\s*DP', filename, re.IGNORECASE)
    if match:
        return f"{match.group(1)} DP"
    return None


def extract_year_from_sheet_name(sheet_name: str) -> int:
    """
    Extrai o ano do nome da sheet
    """
    match = re.search(r'(20\d{2})', str(sheet_name))
    if match:
        return int(match.group(1))
    return None


def normalize_crime_name(crime_raw: str) -> str:
    """
    Normaliza o nome do crime para formato padronizado
    """
    crime_upper = crime_raw.strip().upper()
    
    # Remove números entre parênteses
    crime_upper = re.sub(r'\s*\(\d+\)', '', crime_upper)
    
    # Busca no mapeamento
    for key, value in CRIME_MAPPING.items():
        if key in crime_upper:
            return value
    
    # Se não encontrar, retorna capitalizado
    return crime_raw.strip().title()


def process_excel_file(filepath: str, output_dir: str = "output_json") -> list:
    """
    Processa um arquivo Excel de ocorrências criminais
    Retorna lista de dicionários com os crimes
    """
    filename = os.path.basename(filepath)
    print(f"\n{'='*80}")
    print(f"Processando arquivo: {filename}")
    print(f"{'='*80}")
    
    # Extrai informações do arquivo
    dp_code = extract_dp_from_filename(filename)
    
    if not dp_code:
        print(f"⚠️  AVISO: Não foi possível identificar a DP no arquivo: {filename}")
        return []
    
    if dp_code not in DELEGACIAS_COORDS:
        print(f"⚠️  AVISO: DP {dp_code} não encontrada no dicionário de coordenadas")
        return []
    
    coords = DELEGACIAS_COORDS[dp_code]
    print(f"📍 Delegacia: {dp_code} - {coords['bairro']}")
    print(f"🌍 Coordenadas: ({coords['latitude']}, {coords['longitude']})")
    
    # Lê o arquivo Excel
    try:
        # Tenta ler todas as sheets
        excel_file = pd.ExcelFile(filepath)
        all_crimes = []
        
        for sheet_name in excel_file.sheet_names:
            # Extrai ano da sheet
            year = extract_year_from_sheet_name(sheet_name)
            if not year:
                print(f"\n  ⚠️  Sheet '{sheet_name}' não contém ano válido, pulando...")
                continue
            
            print(f"\n  📄 Processando sheet: {sheet_name} (Ano: {year})")
            df = pd.read_excel(filepath, sheet_name=sheet_name)
            
            # Identifica a coluna de natureza do crime
            crime_column = None
            for col in df.columns:
                if 'natureza' in str(col).lower():
                    crime_column = col
                    break
            
            if crime_column is None:
                print(f"    ⚠️  Coluna 'Natureza' não encontrada na sheet {sheet_name}")
                continue
            
            # Processa cada linha
            crimes_count = 0
            for idx, row in df.iterrows():
                crime_raw = str(row[crime_column]).strip()
                
                # Ignora linhas vazias ou inválidas
                if pd.isna(crime_raw) or crime_raw == '' or crime_raw == 'nan':
                    continue
                
                # Ignora linhas que são contadores de vítimas
                if 'Nº DE VÍTIMAS' in crime_raw.upper() or 'TOTAL DE' in crime_raw.upper():
                    continue
                
                crime_name = normalize_crime_name(crime_raw)
                crime_weight = calculate_weight_crime(crime_name)
                
                # Processa cada mês
                for mes_nome, mes_num in MESES.items():
                    if mes_nome in df.columns:
                        try:
                            quantidade = int(row[mes_nome])
                            
                            # Cria uma ocorrência para cada crime registrado
                            for _ in range(quantidade):
                                crime_entry = {
                                    "crime_name": crime_name,
                                    "crime_weight": crime_weight,
                                    "latitude": str(coords['latitude']),
                                    "longitude": str(coords['longitude']),
                                    "name": coords['bairro'],
                                    "report_date": f"{random.randint(1, 28):02d}/{mes_num:02d}/{year}",
                                    
                                }
                                all_crimes.append(crime_entry)
                                crimes_count += 1
                        except (ValueError, KeyError):
                            continue
            
            print(f"    ✓ {crimes_count} ocorrências processadas")
        
        print(f"\n✅ Total de ocorrências processadas: {len(all_crimes)}")
        
        # Salva JSON individual por DP (todos os anos juntos)
        if all_crimes:
            os.makedirs(output_dir, exist_ok=True)
            output_filename = f"{dp_code.replace(' ', '_')}_crimes.json"
            output_path = os.path.join(output_dir, output_filename)
            
            with open(output_path, 'w', encoding='utf-8') as f:
                json.dump(all_crimes, f, ensure_ascii=False, indent=2)
            
            print(f"💾 Arquivo salvo: {output_path}")
        
        return all_crimes
        
    except Exception as e:
        print(f"❌ ERRO ao processar arquivo: {str(e)}")
        import traceback
        traceback.print_exc()
        return []


def process_multiple_files(input_dir: str = ".", output_dir: str = "output_json"):
    """
    Processa múltiplos arquivos Excel em um diretório
    """
    print("\n" + "="*80)
    print("PROCESSAMENTO DE DADOS CRIMINAIS - CAMPINAS/SP")
    print("="*80)
    
    # Encontra todos os arquivos Excel
    excel_files = []
    for ext in ['*.xlsx', '*.xls']:
        excel_files.extend(Path(input_dir).glob(ext))
    
    if not excel_files:
        print(f"\n⚠️  Nenhum arquivo Excel encontrado em: {input_dir}")
        return
    
    print(f"\n📂 Encontrados {len(excel_files)} arquivo(s) Excel")
    
    all_crimes_combined = []
    
    # Processa cada arquivo
    for filepath in excel_files:
        crimes = process_excel_file(str(filepath), output_dir)
        all_crimes_combined.extend(crimes)
    
    # Salva JSON consolidado
    if all_crimes_combined:
        consolidated_path = os.path.join(output_dir, "all_crimes_consolidated.json")
        with open(consolidated_path, 'w', encoding='utf-8') as f:
            json.dump(all_crimes_combined, f, ensure_ascii=False, indent=2)
        
        print(f"\n{'='*80}")
        print(f"✅ PROCESSAMENTO CONCLUÍDO")
        print(f"{'='*80}")
        print(f"📊 Total de ocorrências: {len(all_crimes_combined)}")
        print(f"💾 Arquivo consolidado: {consolidated_path}")
        
        # Estatísticas - CORRIGIDO
        crimes_by_type = {}
        crimes_by_dp = {}
        crimes_by_year = {}
        heinous_count = 0
        
        for crime in all_crimes_combined:
            # Por tipo
            crime_type = crime.get('crime_name', 'Desconhecido')
            crimes_by_type[crime_type] = crimes_by_type.get(crime_type, 0) + 1
            
            # Por DP
            dp = crime.get('dp', 'N/A')
            crimes_by_dp[dp] = crimes_by_dp.get(dp, 0) + 1
            
            # Por ano
            year = crime.get('year', 'N/A')
            crimes_by_year[year] = crimes_by_year.get(year, 0) + 1
            
            # Hediondos
            if crime.get('crime_weight', 3) == 9:
                heinous_count += 1
        
        print(f"\n📈 ESTATÍSTICAS:")
        print(f"   • Crimes hediondos (peso 9): {heinous_count}")
        print(f"   • Crimes comuns (peso 3): {len(all_crimes_combined) - heinous_count}")
        print(f"   • Tipos de crime diferentes: {len(crimes_by_type)}")
        print(f"   • Delegacias processadas: {len(crimes_by_dp)}")
        
        print(f"\n📅 OCORRÊNCIAS POR ANO:")
        for year in sorted(crimes_by_year.keys()):
            print(f"   • {year}: {crimes_by_year[year]} ocorrências")
        
        print(f"\n🔝 TOP 5 CRIMES MAIS FREQUENTES:")
        top_crimes = sorted(crimes_by_type.items(), key=lambda x: x[1], reverse=True)[:5]
        for i, (crime, count) in enumerate(top_crimes, 1):
            print(f"   {i}. {crime}: {count} ocorrências")
        
        print(f"\n🚔 OCORRÊNCIAS POR DELEGACIA:")
        for dp in sorted(crimes_by_dp.keys()):
            print(f"   • {dp}: {crimes_by_dp[dp]} ocorrências")
        
        print(f"\n{'='*80}\n")


if __name__ == "__main__":
    # Processa arquivos no diretório atual
    process_multiple_files(input_dir=".", output_dir="output_json")
    
    print("\n💡 INSTRUÇÕES DE USO:")
    print("   1. Coloque todos os arquivos .xlsx das delegacias no mesmo diretório deste script")
    print("   2. Execute: python process_crime_data_fixed.py")
    print("   3. Os JSONs serão gerados na pasta 'output_json/'")
    print("\n   Formato do nome do arquivo esperado:")
    print("   OcorrenciaMensal(Criminal)-[NN] DP - Campinas_[YYYYMMDD]_[HHMMSS].xlsx")
    print("   Cada sheet deve ter o nome do ano (ex: 2022, 2023, 2024, 2025)")
    print("\n" + "="*80)