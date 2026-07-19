# Agent setup

Mica exposes the same incident record used by the web workspace through MCP stdio.

## Start the local stack

```bash
make demo-up
curl -fsS http://127.0.0.1:8787/health
```

The health response must be `{"status":"ok"}` before the agent starts an investigation.

## Guided setup

Give a compatible coding agent this instruction from the Mica repository:

```text
Read and follow http://127.0.0.1:8787/agent-onboarding/SKILL.md
```

The onboarding file verifies daemon health, MCP tools, service context, and the task skill.

## Manual MCP configuration

Run the client from the repository root or configure its working directory as the absolute Mica repository path.

```json
{
  "mcpServers": {
    "mica": {
      "command": "go",
      "args": ["run", "./cmd/mica", "mcp"]
    }
  }
}
```

For live telemetry, pass `MICA_PROMETHEUS_URL` and `MICA_SERVICE_CATALOG` in the MCP client environment. Keep credentials out of configuration committed to Git.

## Continue a human-created incident

1. Open the incident in the workspace.
2. Select **Agent**.
3. Copy **Continue this incident** into the connected agent.
4. Confirm agent calls appear in **Activity**.

The handoff includes the full incident ID, service ID, evidence IDs, and appropriate investigation or recovery skill. It tells the agent not to create a second incident.

## Expected MCP tools

- Read: `get_service_context`, `inspect_service`, `find_correlations`
- Record: `record_skill_run`, `record_hypothesis`, `record_change`
- Controlled workflow: `propose_action`, `prepare_incident_update`, `record_audit_findings`, `verify_recovery`

MCP does not expose deploy, restart, rollback, Prometheus mutation, or external publication operations.
