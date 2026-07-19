---
name: mica-incident-communications
description: Prepare evidence-grounded incident, mitigation, approval, recovery, and handoff updates for Slack, Discord, Telegram, or other configured Mica destinations. Use when a team needs an operational update based on a Mica incident. Do not use for casual chat, marketing copy, or unapproved external publication.
---

# Mica incident communications

Use the persisted Mica incident as the source of truth. The goal is to keep humans aligned without leaking sensitive operational context or turning model prose into an unaudited message.

## Workflow

1. Read the incident, timeline, evidence, hypothesis, action, and verification resources.
2. Select one update type: incident opened, investigation update, approval needed, mitigation applied, recovery verified, or handoff.
3. Separate measured facts, agent hypotheses, human decisions, and unresolved questions.
4. Determine the intended audience and configured destinations.
5. Call `prepare_incident_update` with the incident, update type, audience, and destination IDs.
6. Review the returned facts, redactions, and provider previews.
7. Stop for human approval. Do not imply approval from silence or a previous unrelated approval.
8. After explicit approval, call `publish_incident_update` with the immutable broadcast ID.
9. Report each destination result independently and preserve failed destinations for safe retry.

## Required content

Include only what is relevant:

- incident ID, service, environment, severity, and status
- user or system impact when measured
- strongest evidence with values and timestamps
- current hypothesis clearly labeled as a hypothesis
- current owner or next action
- verification result when one exists
- link to the Mica workspace when reachable

## Safety and redaction

Never include:

- webhook URLs, bot tokens, credentials, or secret values
- raw customer data or payloads
- private source snippets or unapproved repository paths
- full PromQL when audience policy excludes it
- unsupported root-cause or recovery claims

Publishing is an external side effect. Require the exact broadcast, destination set, rendered payload, and human approval required by Mica policy.

## Stop conditions

Stop with `needs_human` when:

- audience or destination is ambiguous
- the incident facts are stale or conflicting
- recovery has not been verified but a recovery message was requested
- redaction policy cannot safely classify included fields
- the prepared message differs from what the human approved

## Final response

Return the broadcast ID, update type, destinations, redactions, approval status, delivery receipts, failed destinations, and safe retry instruction.
