#!/usr/bin/env bash

set -euo pipefail

# Allow the dev-implementation agent (or any pipeline agent) to bypass this hook
# by setting CLAUDE_AGENT_BYPASS=1 in the environment.
if [ "${CLAUDE_AGENT_BYPASS:-0}" = "1" ]; then
  exit 0
fi

input="$(cat)"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

# Parse file_path from the tool input JSON. Best-effort: if no parser is
# available or parsing fails, emit a warning and allow the edit through.
file_path=""
parse_json_path='import sys,json; d=json.load(sys.stdin); print(d.get("file_path",""))'
if command -v python3 >/dev/null 2>&1; then
  file_path="$(printf '%s' "$input" | python3 -c "$parse_json_path" 2>/dev/null)" || true
elif command -v python >/dev/null 2>&1; then
  file_path="$(printf '%s' "$input" | python -c "$parse_json_path" 2>/dev/null)" || true
else
  echo "AVISO: python3/python nao encontrado — verificacao de diretorio protegido ignorada." >&2
  exit 0
fi

if [ -z "$file_path" ]; then
  exit 0
fi

protected_dirs=(
  "$repo_root/internal/"
  "$repo_root/cmd/"
  "$repo_root/migrations/"
)

for dir in "${protected_dirs[@]}"; do
  if [[ "$file_path" == "$dir"* ]]; then
    echo "BLOQUEADO: tentativa de editar '$file_path' diretamente." >&2
    echo "" >&2
    echo "Regra do repo: iniciar pelo fluxo /triage -> /guardrails -> /dev -> /code-review -> /qa." >&2
    echo "Use os agents e o report da delivery antes de editar codigo sensivel." >&2
    exit 2
  fi
done

exit 0
