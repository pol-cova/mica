---
name: mica-verify-recovery
description: Verify whether a completed code or operational change recovered a Mica incident using fresh telemetry and the incident's original baseline criteria. Use only after a change and traffic or workload has produced new samples. Do not redefine success criteria after seeing results.
---

# Mica recovery verification

Verify the outcome independently from the implementation and tests. Passing tests are not production recovery.

## Inputs

Require:

- incident ID
- completed change or action summary
- confirmation that the updated service is running
- a verification window with fresh telemetry

## Workflow

1. Read the incident and its original degraded signals, healthy baseline, required signals, and tolerances.
2. Confirm that the change has been recorded with `record_change` or is visible in the incident timeline.
3. Confirm that fresh samples exist after the change timestamp.
4. Call `inspect_service` only when needed to confirm sample freshness or understand missing data.
5. Call `verify_recovery` with the saved incident ID and verification window.
6. Report `recovered`, `partially_recovered`, `unresolved`, or `insufficient_data` exactly as returned by Mica.
7. For partial or failed recovery, identify which original signals remain outside tolerance and recommend the next investigation step.

## Non-negotiable rules

- Use the original baseline, required signals, and tolerances.
- Do not remove a failing signal from verification.
- Do not move the baseline window to make the result pass.
- Do not infer recovery from unit tests, a successful build, or one healthy request.
- Do not close an incident when Mica reports insufficient data.

## Final response

Return:

- incident ID
- verification status
- verification window
- before, degraded, and current value for each required signal
- pass/fail status for each signal
- whether the incident can be closed
- next action when unresolved or insufficient
