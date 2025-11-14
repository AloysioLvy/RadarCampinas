#!/bin/bash
set -e

# ============================================================================
# Script para aplicar migrations SQL manualmente - MySQL Version
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
echo "║     🔧 Radar Campinas - Apply Migrations (MySQL)         ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Valores padrão (podem ser sobrescritos por variáveis de ambiente)
DB_HOST="${DB_HOST:-regulus.cotuca.unicamp.br}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-BD24452}"
DB_PASSWORD="${DB_PASSWORD:-BD24452}"
DB_NAME="${DB_NAME:-BD24452}"

# Caminho do arquivo de migrations
MIGRATIONS_FILE="./backend/internal/database/migrations/knowledge_base_schema_mysql.sql"

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
      echo "  --host <HOST>           Host do MySQL (padrão: regulus.cotuca.unicamp.br)"
      echo "  --port <PORT>           Porta do MySQL (padrão: 3306)"
      echo "  --user <USER>           Usuário do MySQL (padrão: BD24452)"
      echo "  --password <PASSWORD>   Senha do MySQL (padrão: BD24452)"
      echo "  --database <DB>         Nome do banco de dados (padrão: BD24452)"
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

# Verificar se mysql está instalado
if ! command -v mysql &> /dev/null; then
  echo -e "${RED}❌ mysql client não está instalado${NC}"
  echo "   Instale o MySQL client:"
  echo "   Ubuntu/Debian: sudo apt-get install mysql-client"
  echo "   MacOS: brew install mysql-client"
  exit 1
fi

# Testar conexão
echo -e "${BLUE}🔍 Testando conexão com o banco de dados...${NC}"

if ! mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1" "$DB_NAME" > /dev/null 2>&1; then
  echo -e "${RED}❌ Não foi possível conectar ao banco de dados${NC}"
  echo "   Verifique as credenciais e se o servidor está acessível"
  exit 1
fi
echo -e "${GREEN}✅ Conexão estabelecida${NC}"
echo ""

# Aplicar migrations
echo -e "${BLUE}🚀 Aplicando migrations...${NC}"
echo ""

if mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < "$MIGRATIONS_FILE"; then
  echo ""
  echo -e "${GREEN}✅ Migrations aplicadas com sucesso!${NC}"
  echo ""
  
  # Verificar tabelas criadas
  echo -e "${BLUE}🔍 Verificando tabelas criadas...${NC}"
  
  TABLES=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -N -e "
    SELECT TABLE_NAME 
    FROM information_schema.TABLES 
    WHERE TABLE_SCHEMA = '$DB_NAME'
    AND (TABLE_NAME LIKE 'curated_%' 
      OR TABLE_NAME LIKE 'external_%'
      OR TABLE_NAME LIKE 'features_%'
      OR TABLE_NAME LIKE 'analytics_%')
    ORDER BY TABLE_NAME
  ")
  
  if [ -z "$TABLES" ]; then
    echo -e "${YELLOW}⚠️  Nenhuma tabela encontrada${NC}"
  else
    echo "$TABLES" | while read -r table; do
      if [ -n "$table" ]; then
        echo -e "   ${GREEN}✓${NC} Tabela: $table"
      fi
    done
  fi
  echo ""
  
  # Verificar feriados inseridos
  echo -e "${BLUE}🔍 Verificando feriados inseridos...${NC}"
  HOLIDAY_COUNT=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -N -e "SELECT COUNT(*) FROM external_holidays")
  echo -e "   ${GREEN}✓${NC} Total de feriados: $HOLIDAY_COUNT"
  echo ""
  
  # Verificar versão da migration
  echo -e "${BLUE}🔍 Verificando versão da migration...${NC}"
  MIGRATION_VERSION=$(mysql -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -N -e "SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1" 2>/dev/null || echo "N/A")
  echo -e "   ${GREEN}✓${NC} Versão: $MIGRATION_VERSION"
  echo ""
  
  echo -e "${GREEN}🎉 Processo concluído!${NC}"
  exit 0
else
  echo ""
  echo -e "${RED}❌ Erro ao aplicar migrations${NC}"
  echo "   Verifique os logs acima para mais detalhes"
  exit 1
fi