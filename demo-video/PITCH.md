# Demo pitch

Coding agents understand repositories. During a production incident, they are still missing the live context: what changed, what healthy looked like, and whether their fix actually worked. Mica closes that gap.

Mica is a local production engineering workspace shared by humans and coding agents. A Go daemon reads Prometheus, stores one incident in SQLite, serves the React workspace, and exposes the same record through typed MCP tools.

In the demo, a real checkout service produces Prometheus metrics and an N+1 query regression is triggered. Mica compares an aligned healthy baseline with the incident window. P95 latency climbs from 47.5 to 142.8 milliseconds, and database queries per request rise from 2 to 4.2.

Each signal keeps its evidence ID, values, threshold, time windows, query, and provenance. The Agent tab creates an incident-specific handoff for Codex. GPT-5.6 uses typed MCP tools to inspect the service, test hypotheses, and record the exact change, files, and tests back into the incident.

The human and agent share one timeline. Deployment, restart, rollback, and external publication stay outside the tool surface. Recovery is checked against the original baseline and degraded signals, so the agent cannot move the goalposts.

Codex accelerated the Go daemon, React workspace, MCP tools, reusable agent skills, Docker demo, tests, docs, and this Remotion film. GPT-5.6 helped choose the local-daemon architecture and keep evidence, approval, and verification as product boundaries.

Mica turns agent reasoning into reviewable production engineering—from regression to verified recovery. And yes, this demo was built, narrated, and recorded with GPT-5.6-Sol in Codex.
