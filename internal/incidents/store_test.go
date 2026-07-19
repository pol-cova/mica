package incidents

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRecoveryUsesSavedIncidentSignals(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SetDemoRegression(true); err != nil {
		t.Fatal(err)
	}
	incident, err := store.Detect("checkout", zeroTime(), zeroTime(), zeroTime(), zeroTime())
	if err != nil {
		t.Fatal(err)
	}
	if incident.Evidence[2].Classification != "degraded" {
		t.Fatalf("query evidence = %s, want degraded", incident.Evidence[2].Classification)
	}
	if _, err := store.RecordChange(incident.ID, Change{Summary: "batch checkout item lookup"}); err != nil {
		t.Fatal(err)
	}
	if err := store.SetDemoRegression(false); err != nil {
		t.Fatal(err)
	}
	verified, err := store.Verify(incident.ID)
	if err != nil {
		t.Fatal(err)
	}
	if verified.Status != "resolved" || verified.Verification.Status != "recovered" {
		t.Fatalf("verification = %#v, want recovered/resolved", verified.Verification)
	}
	for _, signal := range verified.Verification.Signals {
		if signal.Classification != "stable" {
			t.Errorf("%s = %s, want stable", signal.Signal, signal.Classification)
		}
	}
}

func TestHypothesisRejectsForeignEvidence(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	incident, err := store.Detect("checkout", zeroTime(), zeroTime(), zeroTime(), zeroTime())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.RecordHypothesis(incident.ID, Hypothesis{Summary: "unsupported", EvidenceIDs: []string{"ev_missing"}}); err == nil {
		t.Fatal("expected foreign evidence to be rejected")
	}
}

func TestProposalRequiresIncidentEvidenceAndCanBeReviewed(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	incident, err := store.Detect("checkout", zeroTime(), zeroTime(), zeroTime(), zeroTime())
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.ProposeAction(incident.ID, ActionProposal{ActionType: "replay_traffic", Summary: "Replay checkout traffic", ExpectedEffect: "Fresh samples", RiskStatement: "Demo only", VerificationPlan: "Compare original signals", EvidenceIDs: []string{"other"}})
	if err == nil {
		t.Fatal("expected foreign evidence to fail")
	}
	proposed, err := store.ProposeAction(incident.ID, ActionProposal{ActionType: "replay_traffic", Summary: "Replay checkout traffic", ExpectedEffect: "Fresh samples", RiskStatement: "Demo only", VerificationPlan: "Compare original signals", EvidenceIDs: []string{incident.Evidence[0].ID}})
	if err != nil {
		t.Fatal(err)
	}
	proposal := proposed.Proposals[0]
	if proposal.Status != "pending_review" || proposal.RiskLevel != "propose" {
		t.Fatalf("proposal = %#v", proposal)
	}
	reviewed, err := store.ReviewProposal(incident.ID, proposal.ID, "reviewed", "Paul", "Safe demo action")
	if err != nil {
		t.Fatal(err)
	}
	if reviewed.Proposals[0].Status != "reviewed" || reviewed.Proposals[0].ReviewedBy != "Paul" {
		t.Fatalf("review = %#v", reviewed.Proposals[0])
	}
	if len(reviewed.Timeline) < 3 {
		t.Fatalf("timeline = %#v", reviewed.Timeline)
	}
}

func TestMarkdownReportSeparatesFactsAndHypotheses(t *testing.T) {
	incident := Incident{ID: "inc_1", ServiceID: "checkout", Status: "investigating", Evidence: []SignalComparison{{ID: "ev_1", Signal: "p95 latency", Unit: "ms", BaselineValue: 120, IncidentValue: 410, PercentDelta: 242, Classification: "degraded"}}, Hypotheses: []Hypothesis{{Summary: "N+1 query loop", Confidence: "medium", EvidenceIDs: []string{"ev_1"}, Alternatives: []string{"database saturation"}, NextStep: "inspect checkout loop"}}}
	report := MarkdownReport(incident)
	if !strings.Contains(report, "## Measured evidence") || !strings.Contains(report, "## Hypotheses (agent-authored)") || !strings.Contains(report, "N+1 query loop") {
		t.Fatalf("report = %s", report)
	}
}

func TestSQLitePersistsIncidentAcrossStoreReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mica.db")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	incident, err := store.Detect("checkout", zeroTime(), zeroTime(), zeroTime(), zeroTime())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	persisted, ok := reopened.Incident(incident.ID)
	if !ok || len(persisted.Evidence) != 3 || len(persisted.Timeline) != 1 {
		t.Fatalf("persisted incident = %#v", persisted)
	}
}

func TestReadRefreshesStateWrittenByAnotherLocalClient(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mica.db")
	workspace, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer workspace.Close()
	agent, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer agent.Close()
	incident, err := agent.Detect("checkout", zeroTime(), zeroTime(), zeroTime(), zeroTime())
	if err != nil {
		t.Fatal(err)
	}
	if refreshed, ok := workspace.Incident(incident.ID); !ok || len(refreshed.Evidence) != 3 {
		t.Fatalf("workspace did not refresh agent state: %#v", refreshed)
	}
}

func TestPreparedUpdateRequiresExplicitApproverForDelivery(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "mica.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	incident, err := store.Detect("checkout", zeroTime(), zeroTime(), zeroTime(), zeroTime())
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := store.PrepareUpdate(incident.ID, IncidentUpdate{UpdateType: "investigation_update", Audience: "engineering", Body: "safe fact", PreparedBy: "agent", Destinations: []DeliveryReceipt{{DestinationID: "slack", Provider: "slack", Status: "pending"}}})
	if err != nil {
		t.Fatal(err)
	}
	update := prepared.Updates[0]
	if _, err := store.RecordUpdateDelivery(incident.ID, update.ID, "", update.Destinations); err == nil {
		t.Fatal("expected approval requirement")
	}
	delivered, err := store.RecordUpdateDelivery(incident.ID, update.ID, "Paul", []DeliveryReceipt{{DestinationID: "slack", Provider: "slack", Status: "delivered", Attempts: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if delivered.Updates[0].Status != "delivered" || delivered.Updates[0].ApprovedBy != "Paul" {
		t.Fatalf("update = %#v", delivered.Updates[0])
	}
}

func TestPreparedUpdateCanRemainLocalOnly(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "mica.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	incident, err := store.Detect("checkout", zeroTime(), zeroTime(), zeroTime(), zeroTime())
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := store.PrepareUpdate(incident.ID, IncidentUpdate{UpdateType: "investigation_update", Audience: "engineering", Body: "safe fact", PreparedBy: "agent"})
	if err != nil {
		t.Fatal(err)
	}
	if prepared.Updates[0].Status != "local_only" || len(prepared.Updates[0].Destinations) != 0 {
		t.Fatalf("update = %#v", prepared.Updates[0])
	}
}

func TestHighSeverityAuditFindingRequiresEvidenceAndVerification(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "mica.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	incident, err := store.Detect("checkout", zeroTime(), zeroTime(), zeroTime(), zeroTime())
	if err != nil {
		t.Fatal(err)
	}
	base := AuditFinding{Category: "observability", Severity: "high", AffectedAsset: "checkout", Assessment: "No SLO configured", Risk: "Slow detection", Recommendation: "Add latency SLO", VerificationMethod: "Review configured SLO"}
	if _, err := store.RecordAuditFindings(incident.ID, []AuditFinding{base}, "agent"); err == nil {
		t.Fatal("expected evidence requirement")
	}
	base.EvidenceRefs = []string{"mica://services/checkout"}
	recorded, err := store.RecordAuditFindings(incident.ID, []AuditFinding{base}, "agent")
	if err != nil {
		t.Fatal(err)
	}
	if len(recorded.AuditFindings) != 1 || recorded.AuditFindings[0].Status != "open" {
		t.Fatalf("findings = %#v", recorded.AuditFindings)
	}
}

func zeroTime() (value time.Time) { return }
