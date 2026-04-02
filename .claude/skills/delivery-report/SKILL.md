---
name: delivery-report
description: Create or update a delivery report from the standard template at docs-ai/deliveries/_template/report.md. Use when starting a new delivery or when an agent phase completes and the report needs updating.
---

# Delivery Report Manager

Creates and updates delivery reports for the Expense Tracker pipeline.

## Usage

### Create New Report
1. Read the template at `docs-ai/deliveries/_template/report.md`
2. Create `docs-ai/deliveries/<DELIVERY_ID>/report.md` from the template
3. Fill in available information from the current agent phase
4. Register the delivery in `docs-ai/deliveries/INDEX.md`

### Update Existing Report
1. Read the existing report at `docs-ai/deliveries/<DELIVERY_ID>/report.md`
2. Update the section corresponding to the current agent phase
3. Mark completed sections and add evidence (test results, file paths)

## DELIVERY_ID Prefixes

| Module | Prefix |
|--------|--------|
| Expense Management | `EXPENSE` |
| Auth / Users | `AUTH` |
| Webhooks | `WEBHOOK` |
| Notifications | `NOTIF` |
| Domain / Entities | `DOMAIN` |
| Infra (config, DB, migrations, Docker) | `INFRA` |

## Rules

- Always check `docs-ai/deliveries/INDEX.md` for the next available number.
- Never leave critical fields empty — use `TODO` or `PENDING` markers.
- Update the report incrementally as each agent phase completes.
