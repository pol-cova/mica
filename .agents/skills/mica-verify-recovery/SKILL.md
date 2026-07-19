---
name: mica-verify-recovery
version: 0.1.0
description: Verify whether a completed change recovered a Mica incident using fresh telemetry and the incident's original baseline criteria. Use only after a recorded change and fresh samples. Do not redefine success criteria after seeing results.
---

# Mica recovery verification

Require an incident ID, completed change summary, confirmation that the updated service is running, and a fresh verification window.

1. Call `record_skill_run` with `mica-verify-recovery` and version `0.1.0`, then read the incident’s original degraded signals, healthy baseline, required signals, and tolerances.
2. Confirm the change is recorded and samples are newer than that change.
3. Use `inspect_service` only for freshness or missing-data diagnosis, then call `verify_recovery` with the saved incident and verification window.
4. Return Mica’s exact status: `recovered`, `partially_recovered`, `unresolved`, or `insufficient_data`. For a non-recovery, identify the original signals outside tolerance and recommend the next investigation step.

Never remove a failing signal, move the baseline, infer recovery from tests/builds/a single request, or close an incident with insufficient data.

Return the incident ID, status/window, before/degraded/current values and result for every required signal, closure eligibility, and next action.
