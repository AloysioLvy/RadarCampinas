#!/bin/bash
set -e

# ============================================================================
# Script para executar geração da base de conhecimento manualmente
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
echo "║     🔮 Radar Campinas - Knowledge Base Generator         ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Valores padrão
DAYS_BACK=365
CELL_RESOLUTION=500
API_URL="http://localhost:8080/api/v1/knowledge-base/generate"

# Parse argumentos
while [[ $# -gt 0 ]]; do
  case $1 in
    --days-back)
      DAYS_BACK="$2"
      shift 2
      ;;
    --cell-resolution)
      CELL_RESOLUTION="$2"
      shift 2
      ;;
    --api-url)
      API_URL="$2"
      shift 2
      ;;
    --help)
      echo "Uso: $0 [OPTIONS]"
      echo ""
      echo "Opções:"
      echo "  --days-back <N>         Número de dias para processar (padrão: 365)"
      echo "  --cell-resolution <M>   Resolução das células em metros: 500 ou 1000 (padrão: 500)"
      echo "  --api-url <URL>         URL da API (padrão: http://localhost:8080/api/v1/knowledge-base/generate)"
      echo "  --help                  Mostra esta mensagem"
      echo ""
      echo "Exemplos:"
      echo "  $0"
      echo "  $0 --days-back=180 --cell-resolution=1000"
      exit 0
      ;;
    *)
      echo -e "${RED}❌ Opção desconhecida: $1${NC}"
      echo "Use --help para ver opções disponíveis"
      exit 1
      ;;
  esac
done

# Validações
if [ "$CELL_RESOLUTION" != "500" ] && [ "$CELL_RESOLUTION" != "1000" ]; then
  echo -e "${RED}❌ cell-resolution deve ser 500 ou 1000${NC}"
  exit 1
fi

if [ "$DAYS_BACK" -lt 1 ]; then
  echo -e "${RED}❌ days-back deve ser maior que 0${NC}"
  exit 1
fi

# Mostrar configuração
echo -e "${YELLOW}⚙️  Configuração:${NC}"
echo "   • Dias para processar: $DAYS_BACK"
echo "   • Resolução das células: ${CELL_RESOLUTION}m"
echo "   • URL da API: $API_URL"
echo ""

# Verificar se servidor está rodando
echo -e "${BLUE}🔍 Verificando conectividade...${NC}"
if ! curl -s -f "$API_URL" > /dev/null 2>&1; then
  # Tentar health check
  HEALTH_URL="${API_URL/generate/health}"
  if ! curl -s -f "$HEALTH_URL" > /dev/null 2>&1; then
    echo -e "${RED}❌ Servidor não está acessível em $API_URL${NC}"
    echo "   Certifique-se de que o servidor está rodando:"
    echo "   go run cmd/server/main.go"
    exit 1
  fi
fi
echo -e "${GREEN}✅ Servidor acessível${NC}"
echo ""

# Construir URL com query params
FULL_URL="${API_URL}?days_back=${DAYS_BACK}&cell_resolution=${CELL_RESOLUTION}"

# Executar geração
echo -e "${BLUE}🚀 Iniciando geração da base de conhecimento...${NC}"
echo ""

# Fazer request e capturar resposta
RESPONSE=$(curl -s -X POST "$FULL_URL" -H "Content-Type: application/json")
EXIT_CODE=$?

if [ $EXIT_CODE -ne 0 ]; then
  echo -e "${RED}❌ Erro na comunicação com a API${NC}"
  exit $EXIT_CODE
fi

# Parse resposta (verificar se tem "success")
if echo "$RESPONSE" | grep -q '"status":"success"'; then
  echo -e "${GREEN}✅ Base de conhecimento gerada com sucesso!${NC}"
  echo ""
  echo "📊 Resposta da API:"
  echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
  echo ""
  echo -e "${GREEN}🎉 Processo concluído!${NC}"
  exit 0
else
  echo -e "${RED}❌ Erro na geração da base de conhecimento${NC}"
  echo ""
  echo "📊 Resposta da API:"
  echo "$RESPONSE" | python3 -m json.tool 2>/dev/null || echo "$RESPONSE"
  echo ""
  exit 1
fi
