---
name: agent-audit
description: Audit the quality of agent outputs in a delivery. Detects hallucination, empty fields, unverified checkboxes, shallow handoffs, and scope creep. Use after PR review failure or periodic quality review.
---

# Agent Auditor

Reviews the output quality of all agents that participated in a delivery.

## Usage

Read the delivery report at `docs-ai/deliveries/<DELIVERY_ID>/report.md` and audit each agent section.

## Checks

### Hallucination Detection
- [ ] File paths mentioned in the report actually exist in the repo
- [ ] Function/class names referenced exist in the codebase
- [ ] Test results cited match actual test output
- [ ] Acceptance criteria reference real behavior, not invented

### Completeness
- [ ] No empty or placeholder fields in any agent section
- [ ] All checkboxes have explicit pass/fail (no unchecked items left ambiguous)
- [ ] Residual risks are specific, not generic boilerplate
- [ ] Routing signals match actual changes (e.g., auth changes flagged as SECURITY)

### Handoff Quality
- [ ] Each handoff has all required fields from `docs-ai/agents/08-HANDOFF-CONTRACT.md`
- [ ] Target agent is correct per the orchestration sequence
- [ ] Summary is specific to the delivery, not copy-pasted boilerplate
- [ ] Files changed list matches actual git diff

### Scope Integrity
- [ ] Implementation stays within the scope defined by PM/Triage
- [ ] No unrequested features or refactors added
- [ ] Out-of-scope items from triage were not accidentally implemented

## Output

For each issue found, report:
- **Agent** — which agent produced the issue
- **Field** — which field or section
- **Issue** — what is wrong (hallucination, empty, shallow, scope creep)
- **Evidence** — how you verified (file check, grep, git diff)
- **Fix** — recommended correction

Final verdict: **PASS** or **NEEDS CORRECTION** with severity (critical/warning).
