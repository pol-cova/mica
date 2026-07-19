---
name: mica-security-posture-audit
description: Perform a defensive security posture review of authorized production-facing code, configuration, dependencies, secrets handling, access boundaries, and operational controls. Do not exploit services, scan arbitrary targets, retrieve secrets, or bypass controls.
---

# Mica security posture audit

This is a defensive review workflow, not penetration testing. Use only authorized repositories and configured Mica sources.

## Review areas

- exposed interfaces and trust boundaries
- authentication and authorization assumptions
- secret storage and accidental disclosure risk
- dependency and supply-chain risk
- transport and data-protection configuration
- unsafe defaults and debug or admin surfaces
- input validation and resource-abuse controls
- logging, auditability, and sensitive-data handling
- deployment permissions and separation of duties
- incident, rollback, and recovery controls

## Workflow

1. Retrieve service context, environment, dependencies, owners, and policies.
2. Confirm the authorized repository and configured metadata sources.
3. Inspect code and configuration without executing untrusted content.
4. Classify each result as confirmed vulnerability, risky configuration, missing control, accepted risk, or unknown.
5. Consider exploitability, exposure, blast radius, and existing mitigations before severity.
6. Provide a safe remediation and verification plan.
7. Record only evidence-backed findings through `record_audit_findings`.

## Prohibited behavior

Do not:

- scan or probe arbitrary external hosts
- exploit a vulnerability or create destructive proof of concept
- retrieve, expose, or validate live credentials
- bypass authentication, authorization, rate limits, or policy
- modify production configuration
- represent the output as a formal compliance certification

## Evidence gates

High or critical severity requires:

- exact affected asset
- direct repository, configuration, dependency, or policy evidence
- realistic exposure path
- impact explanation
- existing mitigation analysis
- safe verification method

## Final response

Return scope, evidence sources, prioritized findings, important unknowns, immediate safeguards, longer-term remediation, and verification steps.
