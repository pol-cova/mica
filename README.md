# Mica

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19-149ECA?logo=react&logoColor=white)](https://react.dev/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-4c8955.svg)](LICENSE)

Mica is a local workspace for investigating production regressions. It compares Prometheus windows and stores evidence, hypotheses, code changes, and recovery checks in SQLite. Coding agents access the same incident record through MCP.

Mica reads telemetry but does not deploy, restart services, or modify infrastructure. Outbound messages are sent only after a named approver confirms the prepared draft.

[Watch the 2:59 product demo](demo-video/out/mica-demo.mp4) or inspect the [Remotion source](demo-video/README.md).

## Built with Codex and GPT-5.6

Mica was developed with Codex using a spec-driven workflow. GPT-5.6 Sol at medium and high reasoning led architecture, implementation, debugging, product review, and final integration. GPT-5.6 Terra subagents handled bounded parallel work such as repository exploration, focused implementation, and verification.

The PRD was converted into explicit workstreams, architecture and workflow diagrams, typed interfaces, implementation tasks, and a tracked status document. Codex then built and refined the Go daemon, React workspace, MCP tools, project skills, tests, deterministic evaluation, documentation, and Remotion demo. Each workstream was checked against the PRD and exercised through the local demo before it was marked complete.

Codex also used Mica as an agent, not only as a coding assistant. It connected through the same onboarding instructions and MCP interface included in this repository, inspected incident evidence, recorded investigation work, and verified recovery from fresh telemetry. This closed the loop between the product specification, its implementation, and its real agent workflow.

## Workflow

1. Compare a healthy baseline with the incident window.
2. Inspect exact values, thresholds, Prometheus queries, and collection times.
3. Record hypotheses, code changes, tests, and human decisions.
4. Check the original degraded signals again using fresh telemetry.

Humans use the web workspace. Agents use typed MCP tools. Both update the same persisted incident and timeline.

## Run the demo

```bash
make demo-up
```

Open [http://127.0.0.1:8787](http://127.0.0.1:8787). After roughly 30 seconds of healthy traffic, trigger the bundled checkout N+1 regression:

```bash
MICA_DEMO_CONTROL_URL=http://127.0.0.1:8081 go run ./cmd/mica demo trigger n-plus-one
```

Select **Compare telemetry** in the workspace to create evidence. Reset the scenario when ready to verify recovery:

```bash
MICA_DEMO_CONTROL_URL=http://127.0.0.1:8081 go run ./cmd/mica demo reset
```

For the complete walkthrough, see [docs/demo.md](docs/demo.md).

## Connect an agent

Give a compatible coding agent this instruction after starting Mica:

```text
Read and follow http://127.0.0.1:8787/agent-onboarding/SKILL.md
```

The setup file checks the daemon, configures MCP, verifies the available tools, and selects the task-specific skill. You can also [read it in the repository](web/public/agent-onboarding/SKILL.md).

To continue an incident created in the workspace, open its **Agent** tab and copy the incident handoff. It includes the service, full incident ID, evidence IDs, and next workflow so the agent updates the existing record.

For manual setup, add Mica to the client's MCP configuration:

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

The MCP server exposes service context, regression detection, evidence inspection, correlation, hypothesis/change recording, proposals, safe communication drafting, audit findings, and recovery verification. It cannot deploy, restart services, mutate infrastructure, or publish externally.

## Documentation

| I want to… | Read |
| --- | --- |
| Run the complete regression walkthrough | [Demo](docs/demo.md) |
| Connect a coding agent | [Agent setup](docs/agent-setup.md) |
| Configure a catalog, Prometheus, or destinations | [Configuration](docs/configuration.md) |
| Understand the shared human and agent model | [Architecture](docs/architecture.md) |
| Review implemented PRD scope | [PRD status](docs/prd-status.md) |
| Watch or rebuild the product film | [Demo film](demo-video/README.md) |
| Browse every guide | [Documentation index](docs/README.md) |

Project-scoped skills live in [`.agents/skills`](.agents/skills).

## Development

```bash
make test
make eval
make build
```

`make eval` runs the deterministic MVP scorecard. `make eval-skills` checks the packaged agent skills only. The implementation status for each PRD workstream is recorded in [docs/prd-status.md](docs/prd-status.md).

Set `MICA_PROMETHEUS_BEARER_TOKEN`, or `MICA_PROMETHEUS_BASIC_USER` plus `MICA_PROMETHEUS_BASIC_PASSWORD`, when connecting to a protected Prometheus source. Mica never returns these credentials through its API or MCP server.

To enable outbound delivery, configure only the daemon with `MICA_SLACK_WEBHOOK_URL`, `MICA_DISCORD_WEBHOOK_URL`, and/or `MICA_TELEGRAM_BOT_TOKEN` plus `MICA_TELEGRAM_CHAT_ID`.

## License

MIT — see [LICENSE](LICENSE).
