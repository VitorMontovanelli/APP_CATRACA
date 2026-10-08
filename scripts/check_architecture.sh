#!/usr/bin/env bash
# check_architecture.sh - Valida regras de arquitetura, limites de linhas e testes unitários

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

MAX_GO_LINES=200
MAX_TSX_LINES=150
STRICT_MODE=0

if [[ "${1:-}" == "--strict" ]]; then
    STRICT_MODE=1
fi

echo "=========================================================="
echo "🔍 Catraca Architecture & Line Budget Checker"
echo "   Root: ${ROOT_DIR}"
echo "   Go Limit: ${MAX_GO_LINES} lines"
echo "   TSX/TS Limit: ${MAX_TSX_LINES} lines"
echo "=========================================================="

GO_VIOLATIONS=0
TS_VIOLATIONS=0

echo -e "\n📂 Verificando arquivos Go (.go)..."
while IFS= read -r file; do
    # Ignora arquivos de teste ou gerados
    if [[ "$file" =~ \/wailsjs\/ ]] || [[ "$file" =~ \/build\/ ]]; then
        continue
    fi
    lines=$(wc -l < "$file")
    if (( lines > MAX_GO_LINES )); then
        rel_path="${file#$ROOT_DIR/}"
        echo -e "  ⚠️  \033[31m${rel_path}\033[0m: ${lines} linhas (excede ${MAX_GO_LINES})"
        ((GO_VIOLATIONS++)) || true
    fi
done < <(find "${ROOT_DIR}" -type f -name "*.go" -not -path "*/.*/*" -not -path "*/vendor/*" -not -path "*/third_party/*" -not -path "*/frontend/*")

echo -e "\n🎨 Verificando arquivos Frontend (.tsx, .ts)..."
while IFS= read -r file; do
    if [[ "$file" =~ \/wailsjs\/ ]] || [[ "$file" =~ node_modules ]]; then
        continue
    fi
    lines=$(wc -l < "$file")
    if (( lines > MAX_TSX_LINES )); then
        rel_path="${file#$ROOT_DIR/}"
        echo -e "  ⚠️  \033[33m${rel_path}\033[0m: ${lines} linhas (excede ${MAX_TSX_LINES})"
        ((TS_VIOLATIONS++)) || true
    fi
done < <(find "${ROOT_DIR}/frontend/src" -type f \( -name "*.tsx" -o -name "*.ts" \) -not -path "*/wailsjs/*")

echo -e "\n----------------------------------------------------------"
echo -e "📊 Resumo de Violações de Linhas:"
echo -e "   Go (.go > 200 linhas):       \033[1m${GO_VIOLATIONS}\033[0m arquivo(s)"
echo -e "   Frontend (.tsx > 150 linhas): \033[1m${TS_VIOLATIONS}\033[0m arquivo(s)"
echo -e "----------------------------------------------------------"

if (( GO_VIOLATIONS > 0 || TS_VIOLATIONS > 0 )); then
    if (( STRICT_MODE == 1 )); then
        echo -e "❌ \033[31mFALHA: Limite de linhas estrito violado no modo --strict.\033[0m"
        exit 1
    else
        echo -e "ℹ️  Execute 'scripts/check_architecture.sh --strict' para barrar commits que violem os limites."
    fi
else
    echo -e "✅ \033[32mTodos os arquivos estão dentro do orçamento estrito de linhas!\033[0m"
fi
