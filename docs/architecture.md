# Mica architecture

Mica is a local daemon with a shared incident model.

```text
Codex / compatible agent ── MCP stdio ─┐
                                      ├── Mica daemon ── read-only Prometheus HTTP API
Human workspace ── HTTP + SSE ─────────┘       │
                                                └── local SQLite incident store
```

The incident domain owns evidence, hypotheses, code changes, proposals, verification, audit findings, communications, and timeline. Prometheus is an adapter behind a narrow metrics-comparison port. The web workspace and MCP server use the same incident methods, so humans and agents see the same persisted facts.

Production telemetry is read-only. External publication is the only implemented side effect: it needs a prepared immutable update, configured destination, explicit human approver, receipt recording, and idempotent retry. Mica does not deploy, restart, roll back, change Prometheus, or mutate infrastructure.
