# Mica agent guide

Use this guide to install, run, test, and connect to Mica from any compatible coding agent. Keep the setup local and do not report success until the required checks pass.

## 1. Get the repository

If the current directory does not contain `go.mod`, `docker-compose.yml`, and `.agents/skills/`, clone Mica and enter the repository:

```bash
git clone https://github.com/pol-cova/mica.git
cd mica
```

Do not overwrite an existing checkout or discard local changes.

## 2. Check the requirements

Mica supports macOS and Linux. Windows users should use WSL2. Confirm these commands are available:

```bash
git --version
go version
docker version
docker compose version
make --version
curl --version
```

Go 1.25 or newer is required. Docker must be running before continuing. Node.js and npm are needed only to rebuild the React interface from source.

## 3. Start and verify Mica

From the repository root:

```bash
make demo-up
curl -fsS http://127.0.0.1:8787/health
```

Wait and retry the health request for up to 60 seconds while the containers start. Continue only when it returns:

```json
{"status":"ok"}
```

The human workspace is available at <http://127.0.0.1:8787/workspace>.

## 4. Run the checks

```bash
make test
make eval
make build
```

`make eval` must pass the deterministic MVP checks for regression detection, incident persistence, human review, delivery approval, and recovery verification. If a command fails, report the exact command and error instead of claiming the setup is complete.

## 5. Exercise the demo

Wait about 30 seconds for a healthy Prometheus baseline, then trigger the bundled checkout regression:

```bash
MICA_DEMO_CONTROL_URL=http://127.0.0.1:8081 go run ./cmd/mica demo trigger n-plus-one
```

Open the workspace and select **Compare telemetry**. Confirm that Mica creates an incident with saved evidence for the degraded signals.

Reset the demo and wait for fresh telemetry before checking recovery:

```bash
MICA_DEMO_CONTROL_URL=http://127.0.0.1:8081 go run ./cmd/mica demo reset
```

## 6. Connect through MCP

Configure the MCP client with the absolute repository path as its working directory when supported:

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

List the available tools and confirm at least these are present:

- `get_service_context`
- `detect_regressions`
- `record_hypothesis`
- `record_change`
- `propose_action`
- `prepare_incident_update`
- `verify_recovery`

Call `get_service_context` for a configured service. Do not report the agent connection as complete until the daemon health check and one Mica tool call both succeed.

When continuing a human-created incident, use the exact handoff from the workspace **Agent** tab. Reuse its service ID, incident ID, evidence IDs, and selected skill; do not create a parallel incident.

## Safety boundaries

- Treat production telemetry as read-only.
- Do not deploy, restart, roll back, modify infrastructure, or publish external messages through Mica.
- Do not invent service IDs, Prometheus URLs, evidence, or credentials.
- Keep bearer tokens, passwords, webhook URLs, and Telegram tokens in environment variables. Never commit or repeat them in chat.
- Preserve unrelated local changes.
- External delivery requires a prepared draft and a named human approver in the web workspace.

## Completion report

Return:

- repository path and current commit
- platform and tool versions
- daemon and Prometheus status
- workspace URL
- test, evaluation, and build results
- MCP connection status and available Mica tools
- service ID, environment, and selected workflow when an investigation was requested
- any missing configuration or exact failure

Stop the demo when it is no longer needed:

```bash
make demo-down
```
