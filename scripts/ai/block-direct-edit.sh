#!/usr/bin/env bash

set -euo pipefail

input="$(cat)"
file_path="$(printf '%s' "$input" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('file_path',''))" 2>/dev/null)"
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

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
