# Mica three-minute demo

This walkthrough demonstrates one deliberately focused workflow: detect, investigate, fix, verify, and communicate a checkout N+1 regression.

## 1. Start healthy

```bash
make demo-up
```

Wait about 30 seconds for Prometheus to collect healthy checkout traffic, then open <http://127.0.0.1:8787/workspace>.

The service context identifies the checkout repository, owner, PostgreSQL dependency, mapped signals, SLO reference, runbook, and safe demo notes.

## 2. Trigger the regression

```bash
MICA_DEMO_CONTROL_URL=http://127.0.0.1:8081 go run ./cmd/mica demo trigger
```

Wait about 20 seconds. The traffic generator now exercises an intentional N+1 path: database queries per checkout increase, latency rises, and periodic simulated failures appear.

## 3. Collect evidence

Select **Compare telemetry**. Mica saves an incident and reports p95 latency, error rate, and database queries per request against a healthy baseline. Every card shows the baseline and incident values, comparison threshold, collection windows, query, and persisted evidence ID.

For an MCP-driven flow, configure:

```json
{
  "mcpServers": {
    "mica": { "command": "go", "args": ["run", "./cmd/mica", "mcp"] }
  }
}
```

Open the **Agent** tab and copy the incident handoff into the connected agent. It contains the service, existing incident ID, evidence IDs, and next skill. The agent continues the same record instead of creating a parallel investigation. Every incident-bound MCP call appears in Activity.

## 4. Record the code handoff

Codex records its hypothesis, the batch-query change, changed files, and tests. The human can add context, inspect evidence, review a proposal to replay traffic, and export the Markdown handoff.

## 5. Verify recovery

After the code fix is deployed in a real integration, or to reset the bundled demo, run:

```bash
MICA_DEMO_CONTROL_URL=http://127.0.0.1:8081 go run ./cmd/mica demo reset
```

Wait for fresh telemetry. `verify_recovery` compares the original saved baseline, degraded signal set, and tolerances; it cannot redefine success after seeing results.

## 6. Communicate safely

Configure one or more destinations with daemon-only environment variables:

```bash
MICA_SLACK_WEBHOOK_URL=...
MICA_DISCORD_WEBHOOK_URL=...
MICA_TELEGRAM_BOT_TOKEN=...
MICA_TELEGRAM_CHAT_ID=...
```

Codex may prepare a redacted, persisted update. In the workspace, the engineer reviews the exact preview and enters their name to explicitly approve publication. Mica records independent delivery receipts and retries only failed destinations.

## Validation commands

```bash
make test
make eval
```

Use `make demo-down` to remove the demo containers and volume.
