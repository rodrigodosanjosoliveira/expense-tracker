---
name: validate-delivery
description: Full delivery validation by running scripts/ai/validate-delivery.sh and interpreting results. Use as the final gate before opening a Merge Request.
---

# Validate Delivery

Runs the full automated validation script and interprets the results.

## Usage

1. Run `scripts/ai/validate-delivery.sh <DELIVERY_ID>`
2. Parse the output for errors and warnings
3. For each error, explain what is missing and how to fix it
4. For each warning, assess if it is a blocker or acceptable

## Validation Checks (performed by the script)

- Report file exists at `docs-ai/deliveries/<DELIVERY_ID>/report.md`
- Required sections are present (`PM/Triage`, `Architecture/Guardrails`, `Dev`, `QA`, `Docs`, `Final Checklist`)
- Required fields within each section are non-empty and not placeholder values
- Final Checklist has no unchecked items
- Front matter flags (`require_security`, `require_release_ops`, `arquitetura_change`) are valid booleans
- When `require_security: true`, the `Security` section with `Resultado` field is required
- When `require_release_ops: true`, the `Release/Ops` section with `Resultado` and `Plano de rollback` fields is required
- When `arquitetura_change: true`, `rodrigo_approval` must be filled and not "N/A"

## Output

- **PASS** — delivery is validated, ready for MR
- **FAIL** — list each failure with remediation steps
