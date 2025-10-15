#!/bin/bash
set -e

# ============================================================================
# Script para aplicar migrations SQL manualmente
# ============================================================================

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Banner
echo -e "${BLUE}"
echo "╔═══════════════════════════════════════════════════════════╗"
echo "║     🔧 Radar Campinas - Apply Migrations                 ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Valores padrão (podem ser sobrescritos por variáveis de ambiente)
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-}"
DB_NAME="${DB_NAME:-radar_campinas}"

# Caminho do arquivo de migrations
MIGRATIONS_FILE="internal/database/migrations/knowledge_base_schema.sql"

# Parse argumentos
while [[ $# -gt 0 ]]; do
  case $1 in
    --host)
      DB_HOST="$2"
      shift 2
      ;;
    --port)
      DB_PORT="$2"
      shift 2
      ;;
    --user)
      DB_USER="$2"
      shift 2
      ;;
    --password)
      DB_PASSWORD="$2"
      shift 2
      ;;
    --database)
      DB_NAME="$2"
      shift 2
      ;;
    --file)
      MIGRATIONS_FILE="$2"
      shift 2
      ;;
    --help)
      echo "Uso: $0 [OPTIONS]"
      echo ""
      echo "Opções:"
      echo "  --host <HOST>           Host do PostgreSQL (padrão: localhost)"
      echo "  --port <PORT>           Porta do PostgreSQL (padrão: 5432)"
      echo "  --user <USER>           Usuário do PostgreSQL (padrão: postgres)"
      echo "  --password <PASSWORD>   Senha do PostgreSQL"
      echo "  --database <DB>         Nome do banco de dados (padrão: radar_campinas)"
      echo "  --file <PATH>           Caminho do arquivo de migrations"
      echo "  --help                  Mostra esta mensagem"
      echo ""
      echo "Variáveis de ambiente:"
      echo "  DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME"
      echo ""
      echo "Exemplos:"
      echo "  $0"
      echo "  $0 --host=db.example.com --user=admin --password=secret"
      echo "  DB_PASSWORD=secret $0"
      exit 0
      ;;
    *)
      echo -e "${RED}❌ Opção desconhecida: $1${NC}"
      echo "Use --help para ver opções disponíveis"
      exit 1
      ;;
  esac
done

# Verificar se arquivo de migrations existe
if [ ! -f "$MIGRATIONS_FILE" ]; then
  echo -e "${RED}❌ Arquivo de migrations não encontrado: $MIGRATIONS_FILE${NC}"
  exit 1
fi

# Mostrar configuração (sem senha)
echo -e "${YELLOW}⚙️  Configuração:${NC}"
echo "   • Host: $DB_HOST"
echo "   • Porta: $DB_PORT"
echo "   • Usuário: $DB_USER"
echo "   • Banco: $DB_NAME"
echo "   • Arquivo: $MIGRATIONS_FILE"
echo ""

# Verificar se psql está instalado
if ! command -v psql &> /dev/null; then
  echo -e "${RED}❌ psql não está instalado${NC}"
  echo "   Instale o PostgreSQL client:"
  echo "   Ubuntu/Debian: sudo apt-get install postgresql-client"
  echo "   MacOS: brew install postgresql"
  exit 1
fi

# Testar conexão
echo -e "${BLUE}🔍 Testando conexão com o banco de dados...${NC}"
if [ -n "$DB_PASSWORD" ]; then
  export PGPASSWORD="$DB_PASSWORD"
fi

if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT 1" > /dev/null 2>&1; then
  echo -e "${RED}❌ Não foi possível conectar ao banco de dados${NC}"
  echo "   Verifique as credenciais e se o servidor está acessível"
  exit 1
fi
echo -e "${GREEN}✅ Conexão estabelecida${NC}"
echo ""

# Verificar extensão PostGIS
echo -e "${BLUE}🔍 Verificando extensão PostGIS...${NC}"
POSTGIS_EXISTS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT COUNT(*) FROM pg_extension WHERE extname='postgis'")
if [ "$POSTGIS_EXISTS" -eq "0" ]; then
  echo -e "${YELLOW}⚠️  PostGIS não está instalado${NC}"
  echo "   As migrations tentarão instalar automaticamente"
else
  echo -e "${GREEN}✅ PostGIS já instalado${NC}"
fi
echo ""

# Aplicar migrations
echo -e "${BLUE}🚀 Aplicando migrations...${NC}"
echo ""

if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$MIGRATIONS_FILE"; then
  echo ""
  echo -e "${GREEN}✅ Migrations aplicadas com sucesso!${NC}"
  echo ""
  
  # Verificar schemas criados
  echo -e "${BLUE}🔍 Verificando schemas criados...${NC}"
  SCHEMAS=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT nspname FROM pg_namespace WHERE nspname IN ('curated', 'external', 'features', 'analytics') ORDER BY nspname")
  
  echo "$SCHEMAS" | while read -r schema; do
    if [ -n "$schema" ]; then
      echo -e "   ${GREEN}✓${NC} Schema: $schema"
    fi
  done
  echo ""
  
  # Verificar tabelas criadas
  echo -e "${BLUE}🔍 Verificando tabelas criadas...${NC}"
  TABLES=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -tAc "SELECT schemaname || '.' || tablename FROM pg_tables WHERE schemaname IN ('curated', 'external', 'features', 'analytics') ORDER BY schemaname, tablename")
  
  echo "$TABLES" | while read -r table; do
    if [ -n "$table" ]; then
      echo -e "   ${GREEN}✓${NC} Tabela: $table"
    fi
  done
  echo ""
  
  echo -e "${GREEN}🎉 Processo concluído!${NC}"
  exit 0
else
  echo ""
  echo -e "${RED}❌ Erro ao aplicar migrations${NC}"
  echo "   Verifique os logs acima para mais detalhes"
  exit 1
fi
