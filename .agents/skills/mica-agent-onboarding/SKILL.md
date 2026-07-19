---
name: mica-agent-onboarding
description: Connect a repository and compatible coding agent to a local Mica daemon through MCP. Use when setting up Mica for the first time, adding Mica to an MCP client, checking that the Mica tool surface is available, or preparing an agent to investigate and verify production regressions.
---

# Mica agent onboarding

Connect the current repository to Mica, verify the local runtime, and leave production access read-only.

## 1. Check the repository

Require `go.mod`, `docker-compose.yml`, and `.agents/skills/`. If Mica is not in the current repository, clone it or ask for its absolute path. Do not invent service IDs, Prometheus URLs, or credentials.

## 2. Start Mica

From the Mica repository root, run:

```bash
make demo-up
curl -fsS http://127.0.0.1:8787/health
```

Verify `GET http://127.0.0.1:8787/health` returns `{"status":"ok"}`. Do not continue if the daemon or Prometheus is unavailable.

## 3. Configure MCP

Add this server to the client's MCP configuration, using the absolute repository path as the working directory when the client supports it:

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

For live Prometheus data, pass `MICA_PROMETHEUS_URL` and `MICA_SERVICE_CATALOG` through the client's environment configuration. Keep bearer tokens, basic-auth passwords, webhook URLs, and Telegram tokens in the daemon or client environment; never write them into the repository or conversation.

## 4. Verify the connection

List Mica tools and confirm at least `get_service_context`, `detect_regressions`, `record_hypothesis`, `record_change`, `propose_action`, `prepare_incident_update`, and `verify_recovery` are present. Then call `get_service_context` for a configured service. Stop with the exact error if the service is unknown or context is missing.

Do not start an investigation until the service context and baseline evidence source are available.

## 5. Select the workflow

- Regression investigation: load `mica-investigate-regression`.
- Recovery check: load `mica-verify-recovery`.
- Human handoff: load `mica-incident-handoff`.
- Channel draft: load `mica-incident-communications`.
- Readiness, security, or release review: load the matching audit skill.

Record the selected skill and version with `record_skill_run` after an incident exists.

## 6. Continue an existing incident

When the human supplies an incident handoff from the workspace **Agent** tab, use the exact service ID, incident ID, evidence IDs, and skill in that handoff. Call `inspect_service` for the existing incident. Do not call `detect_regressions` or create a parallel incident unless the human explicitly asks for a new comparison.

Write hypotheses, changes, proposals, findings, drafts, and recovery results to the supplied incident ID. Confirm each saved record appears in the returned incident or timeline.

## Boundaries

Use Mica for telemetry reads and Mica-record mutations only. Do not deploy, restart, roll back, change Prometheus, or publish an update through MCP. External delivery requires a prepared update and a named approver in the web workspace.

## Result

Return the daemon URL, service ID and environment, MCP connection status, available Mica tools, selected skill, and any missing configuration. Do not report setup complete until the health check and one Mica tool call succeed.
