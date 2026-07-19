---
name: mica-release-risk-review
description: Assess operational risk before a release using change scope, service criticality, dependencies, recent incidents, tests, observability, rollback readiness, and blast radius. Do not deploy or approve the release automatically.
---

# Mica release risk review

Evaluate whether a proposed release has enough operational protection to proceed. Risk is contextual: a small diff in a critical path may be riskier than a large isolated change.

## Inputs

- service and environment
- release or commit range
- changed files and components
- test evidence
- deployment strategy
- rollback mechanism
- SLOs and critical signals

## Workflow

1. Retrieve service context, dependencies, criticality, recent incidents, SLOs, and runbooks.
2. Inspect the change scope and identify affected runtime paths.
3. Evaluate migration, configuration, dependency, data, concurrency, and compatibility risk.
4. Check test evidence and observability for the affected behavior.
5. Check canary, rollback, feature-flag, and recovery readiness.
6. Assign risk with explicit drivers and uncertainty.
7. Define required safeguards, blocking conditions, and post-release verification.
8. Record findings or an action proposal; do not deploy.

## Required output

- risk level and confidence
- affected services and blast radius
- top risk drivers
- missing or stale evidence
- conditions that should block release
- recommended rollout and canary plan
- rollback trigger
- exact signals and windows for verification

## Stop conditions

Escalate when the release identity, environment, critical dependencies, rollback path, or required telemetry cannot be resolved.
