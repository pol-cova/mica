# Mica PE/SRE Agent Skills

Versioned agent skills that teach Codex how to use the Mica MCP server for safe, evidence-backed production engineering workflows.

## Incident workflow skills

- `mica-investigate-regression`: investigate a production regression and connect telemetry evidence to repository code.
- `mica-verify-recovery`: verify a change using the incident's original baseline and recovery criteria.
- `mica-incident-handoff`: produce a concise, evidence-backed human handoff.
- `mica-incident-communications`: prepare redacted incident updates and publish only after explicit approval.

## Audit and review skills

- `mica-production-readiness-audit`: review ownership, observability, SLOs, runbooks, deployment safety, capacity, and recovery readiness.
- `mica-security-posture-audit`: run a defensive, authorized security posture review with evidence and safe verification.
- `mica-release-risk-review`: assess release risk, blockers, safeguards, canary strategy, rollback triggers, and verification.

## Install in a repository

Copy the skill directories into:

```text
<repo>/.agents/skills/
```

The skills expect an MCP server named `mica` exposing the tools and resources documented in the Mica PRD.

## Design rules

- Tools provide capabilities; skills provide disciplined workflows.
- Mica remains the source of telemetry facts, policy, state, approval, and verification.
- High-confidence conclusions require evidence.
- External publication and operational mutation require the approval configured by Mica policy.
- Missing sources produce `unknown`, not invented failures.
