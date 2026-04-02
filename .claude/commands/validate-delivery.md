---
description: Validate a delivery report for completeness before opening MR
---

Run `scripts/ai/validate-delivery.sh` on the specified delivery:

```bash
scripts/ai/validate-delivery.sh $ARGUMENTS
```

If the script reports errors, list each one and suggest fixes. If all checks pass, confirm the delivery is ready for MR.
