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
- Required sections are present and non-empty
- Acceptance criteria have explicit status
- DELIVERY_ID is registered in `docs-ai/deliveries/INDEX.md`
- No unresolved `ASSUNCAO` items without escalation note
- Handoff contracts reference valid agent names

## Output

- **PASS** — delivery is validated, ready for MR
- **FAIL** — list each failure with remediation steps
