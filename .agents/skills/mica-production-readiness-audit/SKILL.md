---
name: mica-production-readiness-audit
version: 0.1.0
description: Review service operational readiness from configured Mica context and authorized repository evidence. Use before launch, handoff, migration, or a major release. Do not use as a compliance certification or deployment approval.
---

# Mica production-readiness audit

Review ownership, dependency inventory, telemetry and alerts, SLOs, deploy/canary/rollback safety, runbooks, recovery verification, capacity, backup/recovery, secrets, and documentation freshness.

Retrieve service context first, then record which evidence is configured, unavailable, or stale. A missing source is `unknown`, not automatically a failed control. Inspect only the authorized repository. For every finding, provide category, severity, affected asset, confirmed evidence or explicit unknown state, risk, smallest remediation, owner when known, and verification method. High severity requires direct evidence.

Return readiness status, strongest blockers, important unknowns, prioritized findings, launch safeguards, and the verification plan for every blocker.
