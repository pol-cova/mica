---
name: mica-incident-handoff
description: Create a concise human handoff or incident summary from a Mica incident. Use for on-call handoffs, escalation, status updates, and post-incident summaries. Separate measured facts from agent interpretation and unresolved questions.
---

# Mica incident handoff

Build the handoff from Mica incident resources. Do not rely on conversational memory when a persisted incident record exists.

## Workflow

1. Read `mica://incidents/{incidentId}`.
2. Read `mica://incidents/{incidentId}/timeline`.
3. Retrieve individual evidence resources only when needed to support a key claim.
4. Separate content into:
   - measured facts
   - hypotheses and confidence
   - human decisions
   - code or operational changes
   - verification results
   - unresolved questions
5. Include exact values for the most important impact and recovery signals.
6. Keep the summary concise enough for a new on-call engineer to understand in under two minutes.

## Quality requirements

- Every quantitative claim must come from Mica evidence.
- Clearly label agent-authored hypotheses.
- Do not present an unverified fix as resolved.
- Preserve timestamps for major events.
- State the next owner or required human decision when known.

## Output template

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
