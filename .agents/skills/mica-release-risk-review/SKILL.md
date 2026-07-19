---
name: mica-release-risk-review
version: 0.1.0
description: Assess operational risk before a release using change scope, service criticality, dependencies, incidents, tests, observability, rollback readiness, and blast radius. Do not deploy or approve releases automatically.
---

# Mica release risk review

Require service/environment, release identity, changed components, test evidence, deployment strategy, rollback mechanism, SLOs, and critical signals. Retrieve context; inspect change scope; evaluate migration, configuration, dependency, data, concurrency, and compatibility risk; then check observability, canary, feature-flag, rollback, and recovery readiness.

Return risk level/confidence, affected services/blast radius, risk drivers, missing evidence, release blockers, rollout/canary plan, rollback trigger, and exact verification signals/windows. Escalate when release identity, environment, dependencies, rollback, or required telemetry cannot be resolved. Never deploy.
