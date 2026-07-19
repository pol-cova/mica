# PRD implementation status

This document maps the hackathon PRD to the repository. It describes the implemented MVP, not the post-hackathon roadmap.

| Workstream | Status | Implementation |
| --- | --- | --- |
| Local daemon and storage | Complete | Go daemon, SQLite persistence, HTTP API, and embedded React workspace |
| Service context | Complete | File-backed service catalog with repository, ownership, runtime, dependencies, signals, and runbooks |
| Telemetry comparison | Complete | Read-only Prometheus range queries compare baseline and incident windows and persist evidence |
| Investigation record | Complete | Hypotheses, alternatives, confidence, next steps, code changes, tests, and incident timeline |
| Agent interface | Complete | Typed MCP tools and versioned project skills for investigation, recovery, handoff, communications, readiness, security, and release risk |
| Human controls | Complete | Evidence-backed proposals and explicit review; Mica does not deploy, restart, or mutate infrastructure |
| Recovery | Complete | Fresh telemetry is compared with the incident's saved signals before recovery can be verified |
| Communications | Complete | Redacted local drafts, configured Slack/Discord/Telegram destinations, explicit approval, and delivery receipts |
| Audit findings | Complete | Evidence-linked findings with severity, verification requirements, and timeline attribution |
| Incident report | Complete | Downloadable Markdown separates observations, hypotheses, changes, verification, and communications |
| Deterministic demo | Complete | Docker Compose checkout service, Prometheus, traffic generator, regression trigger/reset/status, and seeded catalog |
| Evaluation | Complete | `make eval` exercises detection, workflow persistence, action review, delivery approval, and recovery verification |

## Product boundaries

The MVP is local-first and read-only for production telemetry. It can prepare actions and external updates, but a person must review them. Autonomous remediation, infrastructure mutation, multi-user access control, hosted operation, traces, logs, and additional telemetry backends remain outside the hackathon scope described by the PRD.
