---
name: delivery-check
description: Quick pre-MR check that validates a delivery report has all required sections filled, acceptance criteria marked, and no empty TODO fields. Lighter than validate-delivery (no script execution).
---

# Delivery Check

Quick validation of a delivery report before opening a Merge Request.

## Usage

Read the report at `docs-ai/deliveries/<DELIVERY_ID>/report.md` and verify:

### Required Sections
- [ ] PM/Triage section filled (demand type, modules, acceptance criteria)
- [ ] Architecture/Guardrails section filled (invariants checked, risks identified)
- [ ] Dev section filled (files changed, technical summary, tests)
- [ ] Code Review section filled (quality, conventions, security surface)
- [ ] QA section filled (acceptance criteria validated, edge cases)
- [ ] Docs section filled (docs updated in same PR)

### Critical Fields
- [ ] DELIVERY_ID follows format `<MODULE>-<NNN>`
- [ ] All acceptance criteria have pass/fail status (no blanks)
- [ ] No unresolved `TODO` or `PENDING` markers in critical fields
- [ ] Routing signals addressed (Security, Release/Ops, Migration if flagged)
- [ ] Residual risks documented or explicitly marked as "none"

### Conditional Sections
- [ ] Security section filled (if routing signal was `SECURITY`)
- [ ] Release/Ops section filled (if routing signal was `RELEASE_OPS`)
- [ ] Migration section filled (if routing signal was `MIGRATION`)
- [ ] Query Performance section filled (if routing signal was `QUERY_PERF`)

## Output

Report one of:
- **READY** — all checks pass, delivery can proceed to MR
- **BLOCKED** — list each missing/incomplete item with file path and line
