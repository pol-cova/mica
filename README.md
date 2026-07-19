# Mica

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![React](https://img.shields.io/badge/React-19-149ECA?logo=react&logoColor=white)](https://react.dev/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![License: MIT](https://img.shields.io/badge/License-MIT-4c8955.svg)](LICENSE)

Mica is a local-first production engineering workspace for finding, explaining, fixing, and verifying production regressions with humans and coding agents.

It combines a Go daemon, a clean incident workspace, Prometheus evidence, a typed MCP server, and versioned PE/SRE agent skills. Production telemetry remains read-only; external communication requires explicit human approval.

## What it does

- Compares a live incident window against a healthy baseline.
- Preserves evidence, hypotheses, code changes, verification, and audit findings in one shared record.
- Gives agents task-level MCP capabilities instead of raw PromQL or infrastructure access.
- Drafts redacted Slack, Discord, and Telegram updates; a person must explicitly approve delivery.
- Ships practical skills for regression investigation, recovery verification, handoff, communications, readiness, security, and release risk reviews.

## Run the demo

```bash
docker compose up --build
```

Open [http://127.0.0.1:8787](http://127.0.0.1:8787). After roughly 30 seconds of healthy traffic, trigger the bundled checkout N+1 regression:

```bash
MICA_DEMO_CONTROL_URL=http://127.0.0.1:8081 go run ./cmd/mica demo trigger n-plus-one
```

Use **Compare** in the workspace to create evidence. Reset the scenario when ready to verify recovery:

```bash
MICA_DEMO_CONTROL_URL=http://127.0.0.1:8081 go run ./cmd/mica demo reset
```

For the complete walkthrough, see [docs/demo.md](docs/demo.md).

## Connect an agent

Add Mica to a compatible MCP client:

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

## Skills

Project-scoped skills live in [`.agents/skills`](.agents/skills):

- `mica-investigate-regression`
- `mica-verify-recovery`
- `mica-incident-handoff`
- `mica-incident-communications`
- `mica-production-readiness-audit`
- `mica-security-posture-audit`
- `mica-release-risk-review`

## Development

```bash
make test
make eval-skills
make build
```

Set `MICA_PROMETHEUS_BEARER_TOKEN`, or `MICA_PROMETHEUS_BASIC_USER` plus `MICA_PROMETHEUS_BASIC_PASSWORD`, when connecting to a protected Prometheus source. Mica never returns these credentials through its API or MCP server.

To enable outbound delivery, configure only the daemon with `MICA_SLACK_WEBHOOK_URL`, `MICA_DISCORD_WEBHOOK_URL`, and/or `MICA_TELEGRAM_BOT_TOKEN` plus `MICA_TELEGRAM_CHAT_ID`.

## License

MIT — see [LICENSE](LICENSE).
