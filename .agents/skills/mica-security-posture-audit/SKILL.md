---
name: mica-security-posture-audit
version: 0.1.0
description: Perform a defensive security posture review of authorized production-facing code, configuration, dependencies, secrets handling, access boundaries, and operational controls. Do not exploit services, scan arbitrary targets, retrieve credentials, or bypass controls.
---

# Mica security posture audit

Retrieve configured context and confirm the authorized repository before inspecting code and configuration without executing untrusted content. Review interfaces, authz assumptions, secret handling, dependencies, transport/data protection, unsafe defaults, input/resource-abuse controls, logging, deployment permissions, and recovery controls.

Classify each outcome as confirmed vulnerability, risky configuration, missing control, accepted risk, or unknown. High or critical severity needs an exact affected asset, direct evidence, realistic exposure path, impact, existing-mitigation analysis, and safe verification method. Never scan arbitrary hosts, create destructive proofs of concept, retrieve or validate live credentials, mutate configuration, or represent the result as compliance certification.

Return scope, evidence sources, prioritized findings, important unknowns, immediate safeguards, remediation, and verification steps.
