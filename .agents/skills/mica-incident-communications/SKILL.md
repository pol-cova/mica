---
name: mica-incident-communications
version: 0.1.0
description: Prepare evidence-grounded incident, mitigation, approval, recovery, and handoff updates for configured Mica destinations. Use when a team needs an operational update based on a persisted Mica incident. Do not use for casual chat or unapproved external publication.
---

# Mica incident communications

Use the persisted incident as the source of truth. Separate measured facts, agent hypotheses, human decisions, and unresolved questions.

1. Read incident evidence, hypothesis, action proposals, changes, and verification.
2. Select one update type: incident opened, investigation update, approval needed, mitigation applied, recovery verified, or handoff.
3. Call `prepare_incident_update` with the incident, audience, and destination IDs.
4. Inspect the returned rendered payload, redactions, and destination list.
5. Stop for explicit human approval. Never infer approval from silence, a previous approval, or a chat message.
6. Human publication must use Mica’s immutable broadcast ID. Report delivery receipts separately and retry only failed destinations.

Never include credentials, webhook URLs, raw customer data, private source snippets, unapproved paths, excluded PromQL, or unsupported cause/recovery claims. A requested recovery update requires persisted Mica recovery verification.

Return the broadcast ID, type, destinations, redactions, approval status, receipts, failed destinations, and safe retry instruction.
