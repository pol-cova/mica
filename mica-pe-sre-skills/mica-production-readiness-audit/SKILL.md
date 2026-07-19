---
name: mica-production-readiness-audit
description: Review whether a service is ready to operate and recover in production using Mica context and repository evidence. Use before launch, handoff, migration, or major release. Do not use as a compliance certification or as permission to deploy.
---

# Mica production-readiness audit

Assess operational readiness from configured evidence. A missing source is `unknown`, not automatically a failed control.

## Review areas

1. Ownership and escalation
2. Service and dependency inventory
3. Metrics, logs, traces, and alert coverage
4. SLOs and error-budget policy
5. Deployment, canary, rollback, and feature-flag safety
6. Runbooks, incident handoff, and recovery verification
7. Capacity, limits, saturation, and dependency assumptions
8. Backup, restoration, data migration, and failure testing where relevant
9. Secret handling, access boundaries, and configuration hygiene
10. Documentation and operational knowledge freshness

## Workflow

1. Call `get_service_context` and retrieve configured repositories, owners, dependencies, SLOs, runbooks, and recent changes.
2. Inspect the repository only within the authorized workspace.
3. Record which evidence sources are configured, unavailable, or stale.
4. Evaluate each review area and collect evidence references.
5. Produce findings using the common `AuditFinding` contract.
6. Prioritize by realistic impact and likelihood, not checklist count.
7. Give every remediation a verification method.
8. Call `record_audit_findings` only after evidence and uncertainty are explicit.

## Finding rules

A finding must include:

- category and severity
- affected asset
- confirmed evidence or an explicit unknown state
- operational risk
- smallest useful remediation
- owner when known
- verification method

High-severity findings require direct evidence. Do not infer a missing control merely because a connector was not configured.

## Final response

Return readiness status, strongest blockers, important unknowns, prioritized findings, launch safeguards, and the verification plan for closing each blocker.
