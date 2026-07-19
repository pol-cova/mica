# Interface language

Mica sounds calm, exact, technical, and brief. It is used during active production work, so measured facts and the next action come before personality.

## Rules

- Lead with the measured result: `2 regressions detected`.
- Show values with their comparison: `47.5 ms → 142.8 ms`.
- Name the window: `Baseline`, `Incident`, or `Latest`.
- Use specific actions: `Compare telemetry`, `Save hypothesis`, `Copy incident handoff`.
- Explain missing math: `Percent change unavailable because the baseline is zero.`
- Do not imply an agent is connected when only the daemon is available.
- Do not call incident data real-time when it comes from a saved comparison window.

## Terminology

| Use | Avoid | Definition |
| --- | --- | --- |
| evidence | insight | Persisted signal comparison with provenance |
| baseline | before | Saved healthy comparison window |
| incident | problem period | Window being compared with the baseline |
| latest | now | Fresh recovery window, only after verification runs |
| hypothesis | root cause | Evidence-linked explanation that is not yet proven |
| recovery verified | fixed | Fresh telemetry meets the original criteria |
| agent handoff | agent prompt | Incident-specific instruction for a connected MCP agent |
