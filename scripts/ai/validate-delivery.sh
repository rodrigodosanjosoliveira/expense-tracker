#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"

if [ $# -lt 1 ]; then
  echo "Uso: scripts/ai/validate-delivery.sh <DELIVERY_ID>" >&2
  echo "Exemplo: scripts/ai/validate-delivery.sh COMP-001" >&2
  exit 1
fi

delivery_id="$1"
report="docs-ai/deliveries/${delivery_id}/report.md"

if [ ! -f "$report" ]; then
  echo "ERRO: report nao encontrado: $report" >&2
  echo "Use o template: docs-ai/deliveries/_template/report.md" >&2
  exit 1
fi

errors=0

meta_block() {
  awk '
    NR==1 && $0=="---" {inside=1; next}
    inside && $0=="---" {exit}
    inside {print}
  ' "$report"
}

get_meta() {
  local key="$1"
  meta_block | sed -n "s/^${key}:[[:space:]]*//p" | head -n1
}

to_bool() {
  local value
  value="$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')"
  case "$value" in
    true|yes|sim|1) echo "true" ;;
    false|no|nao|0|"") echo "false" ;;
    *) echo "invalid" ;;
  esac
}

section_text() {
  local section="$1"
  awk -v section="## ${section}" '
    $0 == section {inside=1; next}
    inside && /^## / {exit}
    inside {print}
  ' "$report"
}

require_section() {
  local section="$1"
  if ! grep -q "^## ${section}$" "$report"; then
    echo "ERRO: secao ausente -> ${section}"
    errors=$((errors + 1))
    return 1
  fi
  return 0
}

invalid_value() {
  local value="$1"
  local normalized
  normalized="$(printf '%s' "$value" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//' | tr '[:upper:]' '[:lower:]')"
  [ -z "$normalized" ] && return 0
  case "$normalized" in
    "<delivery_id>"|"<modulo>"|"todo"|"tbd"|"pendente") return 0 ;;
  esac
  return 1
}

extract_field_value() {
  local section="$1"
  local field="$2"
  local prefix line
  prefix="- ${field}:"

  while IFS= read -r line; do
    case "$line" in
      "${prefix}"*)
        line="${line#"$prefix"}"
        printf '%s\n' "$line" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//'
        return 0
        ;;
    esac
  done < <(section_text "$section")

  printf '%s\n' ""
}

require_field() {
  local section="$1"
  local field="$2"
  local value
  value="$(extract_field_value "$section" "$field")"
  if invalid_value "$value"; then
    echo "ERRO: campo vazio/invalido em ${section} -> ${field}"
    errors=$((errors + 1))
  fi
}

require_field_not_na() {
  local section="$1"
  local field="$2"
  local value normalized
  value="$(extract_field_value "$section" "$field")"
  normalized="$(printf '%s' "$value" | tr '[:upper:]' '[:lower:]' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
  if invalid_value "$value" || [ "$normalized" = "n/a" ]; then
    echo "ERRO: campo obrigatorio sem valor util em ${section} -> ${field}"
    errors=$((errors + 1))
  fi
}

check_mandatory_sections() {
  local sections=(
    "PM/Triage"
    "Architecture/Guardrails"
    "Dev"
    "QA"
    "Docs"
    "Final Checklist"
  )
  for section in "${sections[@]}"; do
    require_section "$section" || true
  done
}

check_mandatory_fields() {
  require_field "PM/Triage" "Escopo fechado"
  require_field "PM/Triage" "Criterios de aceite"
  require_field "PM/Triage" "Riscos"

  require_field "Architecture/Guardrails" "Invariantes aplicaveis"
  require_field "Architecture/Guardrails" "Arquivos provaveis de impacto"
  require_field "Architecture/Guardrails" "Assuncoes"

  require_field "Dev" "Arquivos alterados"
  require_field "Dev" "Resumo tecnico"
  require_field "Dev" "Testes criados/ajustados"
  require_field "Dev" "Evidencias (codigo/teste)"

  require_field "QA" "Resultado"
  require_field "QA" "Cenarios executados"
  require_field "QA" "Regressao validada"

  require_field "Docs" "Docs atualizados"
  require_field "Docs" "Status dos docs"
  require_field "Docs" "Divergencias doc x codigo"
}

check_final_checklist() {
  local checklist
  checklist="$(section_text "Final Checklist")"
  if printf '%s\n' "$checklist" | grep -q '^\- \[ \] '; then
    echo "ERRO: Final Checklist contem itens nao marcados"
    errors=$((errors + 1))
  fi
}

check_flags() {
  local require_front require_security require_release arch_change rodrigo_approval

  require_front="$(to_bool "$(get_meta "require_front_handoff")")"
  require_security="$(to_bool "$(get_meta "require_security")")"
  require_release="$(to_bool "$(get_meta "require_release_ops")")"
  arch_change="$(to_bool "$(get_meta "arquitetura_change")")"
  rodrigo_approval="$(get_meta "rodrigo_approval")"

  for pair in \
    "require_front_handoff:$require_front" \
    "require_security:$require_security" \
    "require_release_ops:$require_release" \
    "arquitetura_change:$arch_change"
  do
    key="${pair%%:*}"
    value="${pair##*:}"
    if [ "$value" = "invalid" ]; then
      echo "ERRO: flag invalida no front matter -> ${key}"
      errors=$((errors + 1))
    fi
  done

  if [ "$require_front" = "true" ]; then
    require_section "FRONT_HANDOFF" || true
    require_field_not_na "FRONT_HANDOFF" "Endpoint/metodo"
    require_field_not_na "FRONT_HANDOFF" "Request/response"
    require_field_not_na "FRONT_HANDOFF" "Status/erros"
  fi

  if [ "$require_security" = "true" ]; then
    require_section "Security" || true
    require_field_not_na "Security" "Resultado"
  fi

  if [ "$require_release" = "true" ]; then
    require_section "Release/Ops" || true
    require_field_not_na "Release/Ops" "Resultado"
    require_field_not_na "Release/Ops" "Plano de rollback"
  fi

  if [ "$arch_change" = "true" ]; then
    if invalid_value "$rodrigo_approval"; then
      echo "ERRO: arquitetura_change=true exige rodrigo_approval preenchido"
      errors=$((errors + 1))
    fi

    local normalized_approval
    normalized_approval="$(printf '%s' "$rodrigo_approval" | tr '[:upper:]' '[:lower:]' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')"
    if [ "$normalized_approval" = "n/a" ] || [ "$normalized_approval" = "nao" ]; then
      echo "ERRO: arquitetura_change=true exige aprovacao explicita do Rodrigo"
      errors=$((errors + 1))
    fi
  fi
}

echo "Validando delivery report: $report"

check_mandatory_sections
check_mandatory_fields
check_final_checklist
check_flags

if [ "$errors" -gt 0 ]; then
  echo "Resultado: FALHOU ($errors erro(s))"
  exit 1
fi

echo "Resultado: OK"
