---
name: mica-incident-handoff
version: 0.1.0
description: Create a concise human handoff or incident summary from a Mica incident. Use for on-call handoffs, escalation, status updates, and post-incident summaries. Separate measured facts from agent interpretation and unresolved questions.
---

# Mica incident handoff

Call `record_skill_run` with `mica-incident-handoff` and version `0.1.0`, then read the persisted incident and timeline first; retrieve individual evidence only for key claims. Separate measured facts, hypotheses/confidence, human decisions, changes, verification, and unresolved questions. Quantitative claims must cite Mica evidence; label hypotheses; never call an unverified fix resolved; preserve major timestamps; and name the next owner or human decision.

Use this structure:

```markdown
## Status
## Impact
## Timeline
## Evidence
## Leading hypothesis
## Changes made
## Verification
## Unresolved questions
## Next action
```
