#!/usr/bin/env bash
# =============================================================================
# setup-github-repo.sh
# Configura boas práticas no repositório rodrigodosanjosoliveira/expense-tracker
# =============================================================================
# Pré-requisitos:
#   - GitHub CLI instalado: https://cli.github.com/
#   - Autenticado: gh auth login
# =============================================================================

set -euo pipefail

# declare -A (associative arrays) requires Bash 4.0+.
# macOS ships Bash 3.2 by default; abort early with a clear message.
if [[ "${BASH_VERSINFO[0]}" -lt 4 ]]; then
  echo "ERRO: Este script requer Bash 4.0 ou superior (versão atual: ${BASH_VERSION})." >&2
  echo "      No macOS, instale uma versão mais recente via Homebrew:" >&2
  echo "        brew install bash" >&2
  echo "      Em seguida, execute o script com: /usr/local/bin/bash \"\$0\"" >&2
  exit 1
fi

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

info()    { echo -e "${GREEN}[INFO]${NC} $1"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $1"; }
error()   { echo -e "${RED}[ERRO]${NC} $1"; exit 1; }

# Repositório alvo:
# - 1º argumento do script: ./setup-github-repo.sh owner/repo
# - ou variável de ambiente: GITHUB_REPO=owner/repo ./setup-github-repo.sh
# - ou auto-detecção via `gh repo view` / `git remote get-url origin`
REPO="${1:-${GITHUB_REPO:-}}"

if [[ -z "${REPO}" ]]; then
  # Tenta detectar via GitHub CLI (repositório atual)
  if gh_repo=$(gh repo view --json nameWithOwner -q .nameWithOwner 2>/dev/null); then
    REPO="${gh_repo}"
  else
    # Fallback: tenta deduzir a partir do remote origin
    if git_remote=$(git remote get-url origin 2>/dev/null); then
      repo_path="${git_remote%.git}"
      repo_path="${repo_path##*:}"          # git@github.com:owner/repo(.git)
      repo_path="${repo_path##github.com/}" # https://github.com/owner/repo(.git)
      REPO="${repo_path}"
    fi
  fi
fi

if [[ -z "${REPO}" ]]; then
  error "Não foi possível determinar o repositório. Informe como argumento (owner/repo) ou defina GITHUB_REPO."
fi

BRANCH="main"

confirm_destructive() {
  warn "Este script fará alterações potencialmente destrutivas no repositório '${REPO}' (proteção de branch, labels, etc.)."
  read -r -p "Tem certeza que deseja continuar? [y/N] " response
  case "${response}" in
    [yY][eE][sS]|[yY])
      ;;
    *)
      info "Operação cancelada pelo usuário."
      exit 0
      ;;
  esac
}

# ---------------------------------------------------------------------------
# 0. Verificações iniciais
# ---------------------------------------------------------------------------
command -v gh &>/dev/null || error "GitHub CLI não encontrado. Instale em: https://cli.github.com/"
gh auth status &>/dev/null || error "Você não está autenticado. Execute: gh auth login"

info "Repositório: $REPO"
info "Branch protegida: $BRANCH"
echo ""

confirm_destructive

# ---------------------------------------------------------------------------
# 1. Proteção da branch main
# ---------------------------------------------------------------------------
info "Configurando proteção da branch '$BRANCH'..."

gh api \
  --method PUT \
  -H "Accept: application/vnd.github+json" \
  "/repos/${REPO}/branches/${BRANCH}/protection" \
  --input - <<EOF
{
  "required_status_checks": {
    "strict": true,
    "contexts": [
      "CI / Lint & Test",
      "CI / Check Conflicts & Up-to-date with main"
    ]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true,
    "require_code_owner_reviews": false,
    "required_approving_review_count": 1,
    "require_last_push_approval": true
  },
  "restrictions": null,
  "allow_force_pushes": false,
  "allow_deletions": false,
  "block_creations": false,
  "required_conversation_resolution": true,
  "lock_branch": false
}
EOF

info "Branch '$BRANCH' protegida com sucesso!"
echo ""

# ---------------------------------------------------------------------------
# 2. Configurações gerais do repositório
# ---------------------------------------------------------------------------
info "Ajustando configurações gerais do repositório..."

gh api \
  --method PATCH \
  -H "Accept: application/vnd.github+json" \
  "/repos/${REPO}" \
  --input - <<EOF
{
  "has_issues": true,
  "has_projects": true,
  "has_wiki": false,
  "allow_squash_merge": true,
  "allow_merge_commit": false,
  "allow_rebase_merge": true,
  "delete_branch_on_merge": true,
  "allow_auto_merge": false,
  "squash_merge_commit_title": "PR_TITLE",
  "squash_merge_commit_message": "PR_BODY"
}
EOF

info "Configurações gerais aplicadas!"
echo ""

# ---------------------------------------------------------------------------
# 3. Labels padronizadas (Conventional Commits + workflow)
# ---------------------------------------------------------------------------
info "Criando labels padronizadas..."

# Remove labels padrão do GitHub para manter tudo limpo
DEFAULT_LABELS=("bug" "documentation" "duplicate" "enhancement" "good first issue" "help wanted" "invalid" "question" "wontfix")
for label in "${DEFAULT_LABELS[@]}"; do
  gh label delete "$label" --repo "$REPO" --yes 2>/dev/null || true
done

# Cria labels organizadas por categoria
declare -A LABELS=(
  # Tipos de mudança (Conventional Commits)
  ["feat"]="0075ca|✨ Nova funcionalidade"
  ["fix"]="d73a4a|🐛 Correção de bug"
  ["docs"]="0052cc|📚 Documentação"
  ["style"]="e4e669|🎨 Estilo / Formatação"
  ["refactor"]="84b6eb|♻️ Refatoração"
  ["perf"]="ff6600|⚡ Performance"
  ["test"]="bfd4f2|🧪 Testes"
  ["chore"]="fef2c0|🔧 Manutenção"
  ["ci"]="c5def5|👷 CI/CD"
  ["build"]="0e8a16|📦 Build"
  # Status / Workflow
  ["wip"]="fbca04|🚧 Em andamento"
  ["needs-review"]="7057ff|👀 Aguardando revisão"
  ["blocked"]="b60205|🚫 Bloqueado"
  ["ready"]="0e8a16|✅ Pronto para merge"
  # Prioridade
  ["priority: high"]="e11d48|🔴 Alta prioridade"
  ["priority: medium"]="f97316|🟠 Média prioridade"
  ["priority: low"]="84cc16|🟢 Baixa prioridade"
  # Breaking change
  ["breaking-change"]="b60205|💥 Breaking Change"
)

for name in "${!LABELS[@]}"; do
  IFS='|' read -r color description <<< "${LABELS[$name]}"
  gh label create "$name" \
    --repo "$REPO" \
    --color "$color" \
    --description "$description" \
    --force 2>/dev/null && echo "  ✓ Label criada: $name" || warn "  Não foi possível criar a label: $name"
done

info "Labels criadas com sucesso!"
echo ""

    # ---------------------------------------------------------------------------
    # 4. Ruleset adicional (GitHub Rulesets - mais moderno que branch protection)
    # ---------------------------------------------------------------------------
    info "Configurando ruleset de proteção da branch main..."

    RULESET_NAME="Protect main branch"

    # Verifica se já existe um ruleset com o mesmo nome/target para tornarmos o script idempotente
    existing_ruleset_id="$(gh api \
      -H "Accept: application/vnd.github+json" \
      "/repos/${REPO}/rulesets?target=branch&per_page=100" \
      --jq ".[] | select(.name==\"${RULESET_NAME}\") | .id" 2>/dev/null | head -n1 || true)"

    if [[ -n "${existing_ruleset_id:-}" ]]; then
      info "Ruleset '${RULESET_NAME}' já existe (id: ${existing_ruleset_id}). Atualizando..."
      RULESET_METHOD="PATCH"
      RULESET_PATH="/repos/${REPO}/rulesets/${existing_ruleset_id}"
    else
      info "Ruleset '${RULESET_NAME}' não encontrado. Criando novo..."
      RULESET_METHOD="POST"
      RULESET_PATH="/repos/${REPO}/rulesets"
    fi

    gh api \
      --method "${RULESET_METHOD}" \
      -H "Accept: application/vnd.github+json" \
      "${RULESET_PATH}" \
      --input - <<EOF
    {
      "name": "Protect main branch",
      "target": "branch",
      "enforcement": "active",
      "conditions": {
        "ref_name": {
          "include": ["refs/heads/main"],
          "exclude": []
        }
      },
      "rules": [
        { "type": "deletion" },
        { "type": "non_fast_forward" },
        { "type": "required_linear_history" },
        {
          "type": "pull_request",
          "parameters": {
            "required_approving_review_count": 1,
            "dismiss_stale_reviews_on_push": true,
            "require_code_owner_review": false,
            "require_last_push_approval": true,
            "required_review_thread_resolution": true
          }
        }
      ]
    }
    EOF

    info "Ruleset configurado!"
    echo ""

# ---------------------------------------------------------------------------
# Resumo
# ---------------------------------------------------------------------------
echo ""
echo "======================================================"
echo -e "${GREEN}✅ Configuração concluída para: $REPO${NC}"
echo "======================================================"
echo ""
echo "O que foi configurado:"
echo "  ✓ Branch 'main' protegida (sem push direto)"
echo "  ✓ Requer PR com 1 aprovação antes do merge"
echo "  ✓ Stale reviews descartados após novo push"
echo "  ✓ CI/CD deve passar antes do merge"
echo "  ✓ Conversas devem ser resolvidas antes do merge"
echo "  ✓ Branch deletada automaticamente após merge"
echo "  ✓ Merge commits desabilitados (squash ou rebase)"
echo "  ✓ Labels padronizadas criadas (Conventional Commits)"
echo "  ✓ Ruleset ativo para linear history"
echo ""
echo "Próximos passos recomendados:"
echo "  1. Adicione o arquivo .github/pull_request_template.md (incluído no pacote)"
echo "  2. Configure seus workflows de CI no .github/workflows/"
echo "  3. Adicione um CODEOWNERS se quiser revisores automáticos"
echo "======================================================"