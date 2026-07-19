---
name: mica-investigate-regression
description: Investigate a production service regression with Mica telemetry, connect evidence to repository code, and record an evidence-backed hypothesis. Use for latency, error-rate, throughput, saturation, or query-amplification regressions. Do not use for routine code debugging without production symptoms or for executing production mutations.
---

# Mica regression investigation

Use Mica as the trusted operational source and the repository as the code source. Do not infer production facts from code alone.

## Inputs

Obtain or resolve:

- affected service
- environment
- reported symptom
- approximate incident window
- repository or working directory

When any input is missing, retrieve service context before asking the user unless Mica reports ambiguity that requires human selection.

## Workflow

1. Call `get_service_context` for the service and environment.
2. Call `detect_regressions` to create or update an incident using a healthy baseline and current window.
3. Inspect the returned degraded signals and evidence IDs.
4. Call `inspect_service` for the smallest set of signals needed to understand the symptom.
5. Call `find_correlations` for the primary degraded signal.
6. Form at least two plausible hypotheses.
7. Search the repository for code paths consistent with the strongest evidence.
8. Distinguish correlation, code evidence, and confirmed causality.
9. Call `record_hypothesis` only when the evidence gates below pass.
10. When a code-level cause is credible, implement the smallest safe fix, add or update a regression test, run relevant tests, and call `record_change`.
11. Do not claim resolution. Hand off to `$mica-verify-recovery` after fresh telemetry exists.

## Evidence gates

A high-confidence hypothesis requires all of the following:

- service context was retrieved
- at least one signal is measured against a saved healthy baseline
- all cited evidence IDs belong to the incident
- at least one alternative explanation was considered
- the repository contains a code path consistent with the observed signal change
- the conclusion states what remains uncertain

When these gates do not pass, record a lower-confidence hypothesis or stop with `needs_human`.

## Stop and escalation conditions

Stop and ask for human input when:

- the service or environment cannot be resolved
- no valid healthy baseline exists
- telemetry is missing or stale
- evidence points to an external dependency not represented in Mica
- the next useful step requires production mutation
- available evidence supports multiple equally plausible causes
- the proposed action exceeds Mica's configured risk ceiling

## Safety rules

- Never invent metric values, deployment events, owners, or runbooks.
- Never treat association scores as proof of causality.
- Never execute deployment, restart, rollback, or infrastructure actions through shell commands as part of this skill.
- Never bypass Mica policy or approval requirements.
- Keep secrets and customer data out of incident notes and model-facing summaries.

## Final response

Return:

- incident ID and current status
- affected service and symptom
- strongest measured changes
- leading hypothesis and confidence
- alternative hypotheses considered
- code evidence and files changed, when applicable
- tests run and results
- unresolved questions
- explicit next step: verify recovery or escalate
