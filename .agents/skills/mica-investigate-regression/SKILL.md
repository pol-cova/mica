---
name: mica-investigate-regression
version: 0.1.0
description: Investigate a production service regression with Mica telemetry, connect evidence to repository code, and record an evidence-backed hypothesis. Use for latency, error-rate, throughput, saturation, or query-amplification regressions. Do not use for routine code debugging without production symptoms or for executing production mutations.
---

# Mica regression investigation

Use Mica as the trusted operational source and the repository as the code source. Do not infer production facts from code alone.

## Required Mica capabilities

`record_skill_run`, `get_service_context`, `detect_regressions`, `inspect_service`, `find_correlations`, `record_hypothesis`, and `record_change`.

## Workflow

1. Once an incident ID exists, call `record_skill_run` with `mica-investigate-regression` and version `0.1.0`. Resolve affected service, environment, reported symptom, incident window, and repository. Retrieve service context before asking for missing information unless Mica reports an ambiguity.
2. Detect or update the incident using its healthy baseline and current window. Inspect degraded evidence IDs and only the signals needed to explain the symptom.
3. Find correlations, form at least two plausible hypotheses, and search the authorized repository for code paths consistent with the evidence.
4. Record a hypothesis only after the gates below pass. If a code cause is credible, implement the smallest safe fix, add a regression test, run relevant tests, and record the change.
5. Do not claim resolution. Hand off to `mica-verify-recovery` after fresh telemetry exists.

## Evidence gates

A high-confidence hypothesis requires retrieved service context, a signal measured against a saved healthy baseline, incident-owned evidence IDs, an alternative explanation, a code path consistent with the observation, and an explicit uncertainty statement. Otherwise use lower confidence or return `needs_human`.

## Safety and stop conditions

Never invent telemetry, owners, deploy events, or runbooks; treat correlation as non-causal; execute no deploy, restart, rollback, or infrastructure action; and never bypass Mica policy. Stop with `needs_human` when service/environment is ambiguous, no baseline exists, telemetry is stale, evidence indicates an unavailable dependency, multiple causes remain equally plausible, or the next action requires production mutation.

## Final response

Return incident ID/status, service/symptom, measured changes, leading and alternative hypotheses, code evidence and changed files, test results, unresolved questions, and the explicit verification or escalation next step.
