# Configuration

## Service catalog

Set `MICA_SERVICE_CATALOG` to a JSON file containing the service ID, name, environment, repository, owner, dependencies, signal mappings, runtime, runbooks, and operational notes. The bundled example is [`demo/mica-services.json`](../demo/mica-services.json).

```bash
MICA_SERVICE_CATALOG=/absolute/path/services.json ./mica serve
```

## Prometheus

```bash
MICA_PROMETHEUS_URL=https://prometheus.example.com ./mica serve
```

Use one authentication method:

| Variable | Purpose |
| --- | --- |
| `MICA_PROMETHEUS_BEARER_TOKEN` | Bearer authentication |
| `MICA_PROMETHEUS_BASIC_USER` | Basic authentication user |
| `MICA_PROMETHEUS_BASIC_PASSWORD` | Basic authentication password |
| `MICA_PROMETHEUS_RATE_WINDOW` | Prometheus range vector duration; defaults to `5m` |

Mica queries the Prometheus HTTP API in read-only mode. Credentials are not returned through HTTP or MCP.

## External destinations

Outbound delivery is disabled unless the daemon receives destination credentials:

| Variable | Destination |
| --- | --- |
| `MICA_SLACK_WEBHOOK_URL` | Slack |
| `MICA_DISCORD_WEBHOOK_URL` | Discord |
| `MICA_TELEGRAM_BOT_TOKEN` | Telegram bot |
| `MICA_TELEGRAM_CHAT_ID` | Telegram chat |

An agent can prepare a draft. A named person must approve the exact draft in the workspace before Mica sends it.
