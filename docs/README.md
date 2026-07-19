# Mica documentation

Start with the task you need to complete.

| Task | Guide |
| --- | --- |
| Run the local checkout regression | [Demo walkthrough](demo.md) |
| Connect Codex or another MCP client | [Agent setup](agent-setup.md) |
| Connect a service catalog and Prometheus | [Configuration](configuration.md) |
| Understand the daemon and shared incident model | [Architecture](architecture.md) |
| Check the implemented hackathon scope | [PRD status](prd-status.md) |
| Write consistent product copy | [Interface language](interface-language.md) |

## Validate the repository

```bash
make test
make eval
make build
```

`make eval` checks regression detection, incident persistence, action review, delivery approval, and recovery verification. `make eval-skills` validates the packaged skill contracts.
