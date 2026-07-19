package incidents

import (
	"fmt"
	"strings"
)

// MarkdownReport renders persisted Mica facts separately from agent-authored
// hypotheses. It is intentionally assembled from the incident record only.
func MarkdownReport(incident Incident) string {
	var report strings.Builder
	fmt.Fprintf(&report, "# Incident %s\n\n", incident.ID)
	fmt.Fprintf(&report, "## Status\n\n%s\n\n", incident.Status)
	fmt.Fprintf(&report, "## Service\n\n%s\n\n", incident.ServiceID)
	report.WriteString("## Measured evidence\n\n")
	for _, evidence := range incident.Evidence {
		fmt.Fprintf(&report, "- **%s:** baseline %.2f %s → incident %.2f %s (%+.1f%%, %s) — evidence `%s`\n", evidence.Signal, evidence.BaselineValue, evidence.Unit, evidence.IncidentValue, evidence.Unit, evidence.PercentDelta, evidence.Classification, evidence.ID)
	}
	report.WriteString("\n## Hypotheses (agent-authored)\n\n")
	if len(incident.Hypotheses) == 0 {
		report.WriteString("No persisted hypothesis.\n")
	}
	for _, hypothesis := range incident.Hypotheses {
		fmt.Fprintf(&report, "- **%s** (confidence: %s; evidence: %s)\n  - Alternatives: %s\n  - Next step: %s\n", hypothesis.Summary, hypothesis.Confidence, strings.Join(hypothesis.EvidenceIDs, ", "), strings.Join(hypothesis.Alternatives, "; "), hypothesis.NextStep)
	}
	report.WriteString("\n## Changes made\n\n")
	if len(incident.Changes) == 0 {
		report.WriteString("No completed code change recorded.\n")
	}
	for _, change := range incident.Changes {
		fmt.Fprintf(&report, "- **%s**\n  - Files: %s\n  - Tests: %s\n", change.Summary, strings.Join(change.Files, ", "), strings.Join(change.Tests, ", "))
	}
	report.WriteString("\n## Action proposals\n\n")
	if len(incident.Proposals) == 0 {
		report.WriteString("No action proposal recorded.\n")
	}
	for _, proposal := range incident.Proposals {
		fmt.Fprintf(&report, "- **%s** (%s, %s)\n  - Expected effect: %s\n  - Risk: %s\n  - Verification: %s\n", proposal.Summary, proposal.Status, proposal.RiskLevel, proposal.ExpectedEffect, proposal.RiskStatement, proposal.VerificationPlan)
	}
	report.WriteString("\n## Verification\n\n")
	if incident.Verification == nil {
		report.WriteString("No recovery verification recorded.\n")
	} else {
		fmt.Fprintf(&report, "**%s** at %s\n\n", incident.Verification.Status, incident.Verification.CheckedAt.Format("2006-01-02 15:04:05 UTC"))
		for _, signal := range incident.Verification.Signals {
			fmt.Fprintf(&report, "- %s: baseline %.2f, degraded %.2f, current %.2f (%s)\n", signal.Signal, signal.BaselineValue, signal.IncidentValue, signal.CurrentValue, signal.Classification)
		}
	}
	report.WriteString("\n## Timeline\n\n")
	for _, event := range incident.Timeline {
		fmt.Fprintf(&report, "- %s — %s (%s)\n", event.OccurredAt.Format("2006-01-02 15:04:05 UTC"), event.Summary, event.ActorName)
	}
	return report.String()
}
