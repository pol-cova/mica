# Mica architecture

Mica is a local daemon with a shared incident model.

```text
Codex / compatible agent ── MCP stdio ─┐
                                      ├── Mica daemon ── read-only Prometheus HTTP API
Human workspace ── HTTP + SSE ─────────┘       │
                                                └── local SQLite incident store
```

The incident domain owns evidence, hypotheses, code changes, proposals, verification, audit findings, communications, and timeline. Prometheus is an adapter behind a narrow metrics-comparison port. The web workspace and MCP server call the same incident methods and write to the same SQLite record.

| Interface | Reads | Writes to Mica |
| --- | --- | --- |
| Human workspace | Service context, signal values, provenance, timeline, recovery | Hypotheses, changes, notes, reviews, approvals |
| MCP agent | Service context, incident evidence, correlations, timeline resources | Skill runs, hypotheses, changes, proposals, drafts, findings, recovery checks |

The workspace **Agent** tab generates a handoff containing the current service, incident ID, evidence IDs, and workflow skill. This keeps both interfaces on the same incident.

Production telemetry is read-only. External publication is the only implemented side effect: it needs a prepared immutable update, configured destination, explicit human approver, receipt recording, and idempotent retry. Mica does not deploy, restart, roll back, change Prometheus, or mutate infrastructure.
