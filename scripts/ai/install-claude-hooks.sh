#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
settings_file="$repo_root/.claude/settings.json"

chmod +x "$repo_root/scripts/ai/block-direct-edit.sh"

cat <<EOF
Repo hook ready:
- script: $repo_root/scripts/ai/block-direct-edit.sh
- settings: $settings_file

If your Claude client does not load repo-level hooks automatically, keep a local/global bridge only as a compatibility layer.
EOF
