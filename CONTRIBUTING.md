# Contributing to Mica

Thanks for contributing. Mica is intentionally narrow: improve the evidence-backed production workflow without widening production permissions.

## Before opening a change

1. Keep telemetry access read-only and external actions explicitly approved.
2. Preserve provenance: claims must be linked to persisted evidence, windows, and values.
3. Add or update tests for behavior changes.
4. Run `make test`, `make eval-skills`, and `npm --prefix web run build`.

## Pull requests

Use a focused title and explain the user impact, safety implications, and verification performed. Do not include secrets, webhook URLs, local databases, generated dependency directories, or production telemetry in a pull request.

## Skills

Skills live under `.agents/skills/`. Each skill needs a clear trigger, tool requirements, evidence gates, safety boundaries, stop conditions, and structured outcome. Run `make eval-skills` after any skill change.
