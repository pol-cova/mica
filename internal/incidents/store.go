package incidents

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mica-dev/mica/internal/storage"
)

type persistedState struct {
	DemoRegression bool       `json:"demoRegression"`
	Services       []Service  `json:"services"`
	Incidents      []Incident `json:"incidents"`
}

type Store struct {
	mu       sync.RWMutex
	path     string
	database *sql.DB
	data     persistedState
	metrics  MetricsSource
	timeline TimelineSink
}

func Open(path string) (*Store, error) {
	s := &Store{path: path}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	database, err := storage.Open(path)
	if err != nil {
		return nil, err
	}
	s.database = database
	contents, exists, err := storage.Load(database)
	if err != nil {
		database.Close()
		return nil, err
	}
	if !exists {
		s.data.Services = []Service{demoService()}
		if err := s.saveLocked(); err != nil {
			database.Close()
			return nil, err
		}
		return s, nil
	}
	if err := json.Unmarshal(contents, &s.data); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if len(s.data.Services) == 0 {
		s.data.Services = []Service{demoService()}
	}
	return s, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.database == nil {
		return nil
	}
	err := s.database.Close()
	s.database = nil
	return err
}

func demoService() Service {
	return Service{ID: "checkout", Name: "checkout", Environment: "demo", Repository: "demo/checkout-service", Owners: []string{"checkout-team"}, Dependencies: []string{"postgres"}, Signals: map[string]string{"p95_latency": "http_request_duration_seconds", "error_rate": "http_requests_total", "queries_per_request": "checkout_db_queries_total"}, Source: SourceMetadata{Kind: "configured demo mapping", RefreshedAt: time.Now().UTC()}}
}

func (s *Store) Services() []Service {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.refreshLocked()
	return append([]Service(nil), s.data.Services...)
}
func (s *Store) Service(id string) (Service, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.refreshLocked()
	for _, v := range s.data.Services {
		if v.ID == id {
			return v, true
		}
	}
	return Service{}, false
}

// Incidents returns the persisted incident record in newest-first order so
// human clients can resume a shared investigation after a page refresh.
func (s *Store) Incidents() []Incident {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.refreshLocked()
	items := append([]Incident(nil), s.data.Incidents...)
	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
	return items
}

func (s *Store) SetDemoRegression(enabled bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.DemoRegression = enabled
	return s.saveLocked()
}

// DemoRegression reports the persisted state of the bundled deterministic
// regression. It does not inspect or change a live service.
func (s *Store) DemoRegression() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.refreshLocked()
	return s.data.DemoRegression
}

// SetMetricsSource swaps the telemetry adapter without changing the incident
// domain. Passing nil restores the bundled deterministic demo source.
func (s *Store) SetMetricsSource(source MetricsSource) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics = source
}

func (s *Store) ReplaceServices(services []Service) error {
	if len(services) == 0 {
		return errors.New("at least one service is required")
	}
	seen := map[string]bool{}
	for _, service := range services {
		if service.ID == "" || service.Name == "" || service.Environment == "" {
			return errors.New("services require id, name, and environment")
		}
		if seen[service.ID] {
			return fmt.Errorf("duplicate service ID %q", service.ID)
		}
		seen[service.ID] = true
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data.Services = append([]Service(nil), services...)
	return s.saveLocked()
}

func (s *Store) SetTimelineSink(sink TimelineSink) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.timeline = sink
}

func (s *Store) Detect(serviceID string, baselineStart, baselineEnd, incidentStart, incidentEnd time.Time) (Incident, error) {
	s.mu.RLock()
	var service Service
	found := false
	for _, candidate := range s.data.Services {
		if candidate.ID == serviceID {
			service, found = candidate, true
		}
	}
	regressed, source := s.data.DemoRegression, s.metrics
	s.mu.RUnlock()
	if !found {
		return Incident{}, fmt.Errorf("service %q not found", serviceID)
	}
	now := time.Now().UTC()
	if baselineStart.IsZero() {
		baselineStart = now.Add(-2 * time.Hour)
	}
	if baselineEnd.IsZero() {
		baselineEnd = now.Add(-time.Hour)
	}
	if incidentStart.IsZero() {
		incidentStart = now.Add(-15 * time.Minute)
	}
	if incidentEnd.IsZero() {
		incidentEnd = now
	}
	evidence := comparisons(regressed)
	if source != nil {
		var err error
		evidence, err = source.Compare(context.Background(), service, baselineStart, baselineEnd, incidentStart, incidentEnd)
		if err != nil {
			return Incident{}, fmt.Errorf("collect telemetry: %w", err)
		}
		if len(evidence) == 0 {
			return Incident{}, errors.New("telemetry source returned no usable signal data")
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	incident := Incident{ID: fmt.Sprintf("inc_%d", now.UnixNano()), ServiceID: serviceID, Status: "investigating", CreatedAt: now, UpdatedAt: now, BaselineStart: baselineStart, BaselineEnd: baselineEnd, IncidentStart: incidentStart, IncidentEnd: incidentEnd, Evidence: evidence}
	s.appendTimeline(&incident, "regression_detected", "Collected baseline comparison evidence", "mica", "Mica")
	s.data.Incidents = append(s.data.Incidents, incident)
	return incident, s.saveLocked()
}

func comparisons(regressed bool) []SignalComparison {
	values := []struct {
		name, unit, query        string
		baseline, bad, tolerance float64
	}{
		{"p95 latency", "ms", "histogram_quantile(0.95, checkout_http_duration_seconds_bucket)", 120, 410, 20},
		{"error rate", "%", "sum(rate(checkout_http_requests_total{code=~\"5..\"}[5m])) / sum(rate(checkout_http_requests_total[5m]))", 0.2, 3.8, 25},
		{"database queries per request", "queries/request", "sum(rate(checkout_db_queries_total[5m])) / sum(rate(checkout_http_requests_total[5m]))", 1.1, 8.3, 20},
	}
	result := make([]SignalComparison, 0, len(values))
	for i, v := range values {
		current := v.baseline
		if regressed {
			current = v.bad
		}
		delta := current - v.baseline
		pct := delta / v.baseline * 100
		class := "stable"
		if pct > v.tolerance {
			class = "degraded"
		}
		result = append(result, SignalComparison{ID: fmt.Sprintf("ev_%d", i+1), Signal: v.name, Unit: v.unit, Query: v.query, BaselineValue: v.baseline, IncidentValue: current, AbsoluteDelta: delta, PercentDelta: pct, Classification: class, TolerancePct: v.tolerance})
	}
	return result
}

func (s *Store) Incident(id string) (Incident, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.refreshLocked()
	for _, v := range s.data.Incidents {
		if v.ID == id {
			return v, true
		}
	}
	return Incident{}, false
}

func (s *Store) refreshLocked() error {
	if s.database == nil {
		return errors.New("store is closed")
	}
	contents, exists, err := storage.Load(s.database)
	if err != nil || !exists {
		return err
	}
	var data persistedState
	if err := json.Unmarshal(contents, &data); err != nil {
		return fmt.Errorf("refresh persisted state: %w", err)
	}
	s.data = data
	return nil
}

func (s *Store) RecordHypothesis(incidentID string, h Hypothesis) (Incident, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Incidents {
		if s.data.Incidents[i].ID == incidentID {
			if !validEvidence(s.data.Incidents[i], h.EvidenceIDs) {
				return Incident{}, errors.New("one or more evidence IDs do not belong to this incident")
			}
			h.ID = fmt.Sprintf("hyp_%d", time.Now().UnixNano())
			h.CreatedAt = time.Now().UTC()
			s.data.Incidents[i].Hypotheses = append(s.data.Incidents[i].Hypotheses, h)
			s.data.Incidents[i].Status = "cause_proposed"
			s.data.Incidents[i].UpdatedAt = h.CreatedAt
			s.appendTimeline(&s.data.Incidents[i], "hypothesis_recorded", h.Summary, "agent", "Mica agent")
			return s.data.Incidents[i], s.saveLocked()
		}
	}
	return Incident{}, errors.New("incident not found")
}

func validEvidence(incident Incident, ids []string) bool {
	if len(ids) == 0 {
		return false
	}
	valid := map[string]bool{}
	for _, evidence := range incident.Evidence {
		valid[evidence.ID] = true
	}
	for _, id := range ids {
		if !valid[id] {
			return false
		}
	}
	return true
}

func (s *Store) ProposeAction(incidentID string, proposal ActionProposal) (Incident, error) {
	if proposal.ActionType == "" || proposal.Summary == "" || proposal.ExpectedEffect == "" || proposal.RiskStatement == "" || proposal.VerificationPlan == "" {
		return Incident{}, errors.New("action type, summary, expected effect, risk statement, and verification plan are required")
	}
	if proposal.ProposedBy == "" {
		proposal.ProposedBy = "Mica agent"
	}
	if proposal.RiskLevel == "" {
		proposal.RiskLevel = "propose"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Incidents {
		if s.data.Incidents[i].ID == incidentID {
			if !validEvidence(s.data.Incidents[i], proposal.EvidenceIDs) {
				return Incident{}, errors.New("one or more evidence IDs do not belong to this incident")
			}
			proposal.ID = fmt.Sprintf("act_%d", time.Now().UnixNano())
			proposal.Status = "pending_review"
			proposal.CreatedAt = time.Now().UTC()
			s.data.Incidents[i].Proposals = append(s.data.Incidents[i].Proposals, proposal)
			s.data.Incidents[i].UpdatedAt = proposal.CreatedAt
			s.appendTimeline(&s.data.Incidents[i], "action_proposed", proposal.Summary, "agent", proposal.ProposedBy)
			return s.data.Incidents[i], s.saveLocked()
		}
	}
	return Incident{}, errors.New("incident not found")
}

func (s *Store) ReviewProposal(incidentID, proposalID, status, reviewer, note string) (Incident, error) {
	if status != "reviewed" && status != "rejected" && status != "deferred" {
		return Incident{}, errors.New("status must be reviewed, rejected, or deferred")
	}
	if reviewer == "" {
		return Incident{}, errors.New("reviewer is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Incidents {
		if s.data.Incidents[i].ID == incidentID {
			for j := range s.data.Incidents[i].Proposals {
				proposal := &s.data.Incidents[i].Proposals[j]
				if proposal.ID == proposalID {
					now := time.Now().UTC()
					proposal.Status, proposal.ReviewedBy, proposal.ReviewNote, proposal.ReviewedAt = status, reviewer, note, &now
					s.data.Incidents[i].UpdatedAt = now
					s.appendTimeline(&s.data.Incidents[i], "proposal_"+status, proposal.Summary, "human", reviewer)
					return s.data.Incidents[i], s.saveLocked()
				}
			}
			return Incident{}, errors.New("action proposal not found")
		}
	}
	return Incident{}, errors.New("incident not found")
}

func (s *Store) PrepareUpdate(incidentID string, update IncidentUpdate) (Incident, error) {
	if update.UpdateType == "" || update.Audience == "" || update.Body == "" {
		return Incident{}, errors.New("update type, audience, and body are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Incidents {
		if s.data.Incidents[i].ID == incidentID {
			update.ID = fmt.Sprintf("upd_%d", time.Now().UnixNano())
			update.Status = "pending_approval"
			if len(update.Destinations) == 0 {
				update.Status = "local_only"
			}
			update.CreatedAt = time.Now().UTC()
			update.IdempotencyKey = fmt.Sprintf("mica-%s", update.ID)
			s.data.Incidents[i].Updates = append(s.data.Incidents[i].Updates, update)
			s.data.Incidents[i].UpdatedAt = update.CreatedAt
			s.appendTimeline(&s.data.Incidents[i], "update_prepared", update.UpdateType+" update prepared", "agent", update.PreparedBy)
			return s.data.Incidents[i], s.saveLocked()
		}
	}
	return Incident{}, errors.New("incident not found")
}

func (s *Store) Update(incidentID, updateID string) (IncidentUpdate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, incident := range s.data.Incidents {
		if incident.ID == incidentID {
			for _, update := range incident.Updates {
				if update.ID == updateID {
					return update, true
				}
			}
		}
	}
	return IncidentUpdate{}, false
}

func (s *Store) RecordUpdateDelivery(incidentID, updateID, approvedBy string, receipts []DeliveryReceipt) (Incident, error) {
	if approvedBy == "" {
		return Incident{}, errors.New("explicit human approver is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Incidents {
		if s.data.Incidents[i].ID == incidentID {
			for j := range s.data.Incidents[i].Updates {
				update := &s.data.Incidents[i].Updates[j]
				if update.ID == updateID {
					if update.Status != "pending_approval" && update.Status != "partially_delivered" {
						return Incident{}, errors.New("update is not pending approval or retry")
					}
					now := time.Now().UTC()
					update.ApprovedBy, update.ApprovedAt, update.Destinations = approvedBy, &now, receipts
					update.Status = "delivered"
					for _, receipt := range receipts {
						if receipt.Status != "delivered" {
							update.Status = "partially_delivered"
							break
						}
					}
					s.data.Incidents[i].UpdatedAt = now
					s.appendTimeline(&s.data.Incidents[i], "update_published", update.UpdateType+" update published: "+update.Status, "human", approvedBy)
					return s.data.Incidents[i], s.saveLocked()
				}
			}
			return Incident{}, errors.New("prepared update not found")
		}
	}
	return Incident{}, errors.New("incident not found")
}

func (s *Store) RecordAuditFindings(incidentID string, findings []AuditFinding, actor string) (Incident, error) {
	if len(findings) == 0 {
		return Incident{}, errors.New("at least one audit finding is required")
	}
	if actor == "" {
		actor = "Mica agent"
	}
	for _, finding := range findings {
		if finding.Category == "" || finding.Severity == "" || finding.AffectedAsset == "" || finding.Assessment == "" || finding.Risk == "" || finding.Recommendation == "" || finding.VerificationMethod == "" {
			return Incident{}, errors.New("audit findings require category, severity, affected asset, assessment, risk, recommendation, and verification method")
		}
		if (finding.Severity == "high" || finding.Severity == "critical") && len(finding.EvidenceRefs) == 0 {
			return Incident{}, errors.New("high and critical findings require evidence references")
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Incidents {
		if s.data.Incidents[i].ID == incidentID {
			now := time.Now().UTC()
			for index := range findings {
				findings[index].ID = fmt.Sprintf("audit_%d_%d", now.UnixNano(), index)
				findings[index].CreatedAt = now
				if findings[index].Status == "" {
					findings[index].Status = "open"
				}
			}
			s.data.Incidents[i].AuditFindings = append(s.data.Incidents[i].AuditFindings, findings...)
			s.data.Incidents[i].UpdatedAt = now
			s.appendTimeline(&s.data.Incidents[i], "audit_findings_recorded", fmt.Sprintf("Recorded %d audit finding(s)", len(findings)), "agent", actor)
			return s.data.Incidents[i], s.saveLocked()
		}
	}
	return Incident{}, errors.New("incident not found")
}

// RecordCapabilityInvocation preserves the agent's read and record actions in
// the shared incident timeline. It deliberately records only a safe tool name,
// never raw arguments that might include sensitive context.
func (s *Store) RecordCapabilityInvocation(incidentID, capability string) error {
	if incidentID == "" || capability == "" {
		return errors.New("incident ID and capability are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Incidents {
		if s.data.Incidents[i].ID == incidentID {
			s.appendTimeline(&s.data.Incidents[i], "capability_invoked", "Agent invoked "+capability, "agent", "Codex via MCP")
			s.data.Incidents[i].UpdatedAt = time.Now().UTC()
			return s.saveLocked()
		}
	}
	return errors.New("incident not found")
}

// RecordSkillInvocation makes the reviewed workflow that guided an agent's
// work visible beside the resulting evidence and records.
func (s *Store) RecordSkillInvocation(incidentID, skillName, version string) error {
	if incidentID == "" || skillName == "" || version == "" {
		return errors.New("incident ID, skill name, and version are required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Incidents {
		if s.data.Incidents[i].ID == incidentID {
			s.appendTimeline(&s.data.Incidents[i], "skill_run", fmt.Sprintf("Guided by %s v%s", skillName, version), "agent", "Codex via MCP")
			s.data.Incidents[i].UpdatedAt = time.Now().UTC()
			return s.saveLocked()
		}
	}
	return errors.New("incident not found")
}

func (s *Store) RecordChange(incidentID string, change Change) (Incident, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Incidents {
		if s.data.Incidents[i].ID == incidentID {
			change.ID = fmt.Sprintf("chg_%d", time.Now().UnixNano())
			change.CreatedAt = time.Now().UTC()
			s.data.Incidents[i].Changes = append(s.data.Incidents[i].Changes, change)
			s.data.Incidents[i].Status = "verifying"
			s.data.Incidents[i].UpdatedAt = change.CreatedAt
			s.appendTimeline(&s.data.Incidents[i], "change_recorded", change.Summary, "agent", "Mica agent")
			return s.data.Incidents[i], s.saveLocked()
		}
	}
	return Incident{}, errors.New("incident not found")
}

func (s *Store) Verify(incidentID string) (Incident, error) {
	return s.VerifyWindow(incidentID, time.Time{}, time.Time{})
}

// VerifyWindow compares fresh post-change telemetry to the incident's saved
// baseline and tolerances. It deliberately never accepts caller-defined success
// thresholds.
func (s *Store) VerifyWindow(incidentID string, verificationStart, verificationEnd time.Time) (Incident, error) {
	s.mu.RLock()
	var stored Incident
	found := false
	for _, candidate := range s.data.Incidents {
		if candidate.ID == incidentID {
			stored, found = candidate, true
			break
		}
	}
	var service Service
	for _, candidate := range s.data.Services {
		if candidate.ID == stored.ServiceID {
			service = candidate
			break
		}
	}
	regressed, source := s.data.DemoRegression, s.metrics
	s.mu.RUnlock()
	if !found {
		return Incident{}, errors.New("incident not found")
	}
	if len(stored.Changes) == 0 {
		return Incident{}, errors.New("record a completed change before verification")
	}
	now := time.Now().UTC()
	if verificationStart.IsZero() {
		verificationStart = now.Add(-5 * time.Minute)
	}
	if verificationEnd.IsZero() {
		verificationEnd = now
	}
	fresh := comparisons(regressed)
	if source != nil {
		var err error
		fresh, err = source.Compare(context.Background(), service, stored.BaselineStart, stored.BaselineEnd, verificationStart, verificationEnd)
		if err != nil {
			return Incident{}, fmt.Errorf("collect verification telemetry: %w", err)
		}
	}
	freshBySignal := make(map[string]SignalComparison, len(fresh))
	for _, signal := range fresh {
		freshBySignal[signal.Signal] = signal
	}
	verifiedSignals := make([]SignalComparison, 0, len(stored.Evidence))
	status := "recovered"
	for _, original := range stored.Evidence {
		current, ok := freshBySignal[original.Signal]
		if !ok {
			original.Classification = "insufficient_data"
			status = "insufficient_data"
			verifiedSignals = append(verifiedSignals, original)
			continue
		}
		original.CurrentValue = current.IncidentValue
		original.AbsoluteDelta = original.CurrentValue - original.BaselineValue
		if original.BaselineValue == 0 {
			original.Classification = "insufficient_data"
			status = "insufficient_data"
		} else {
			original.PercentDelta = original.AbsoluteDelta / original.BaselineValue * 100
			if math.Abs(original.PercentDelta) <= original.TolerancePct {
				original.Classification = "stable"
			} else {
				original.Classification = "degraded"
				if status == "recovered" {
					status = "unresolved"
				}
			}
		}
		verifiedSignals = append(verifiedSignals, original)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Incidents {
		if s.data.Incidents[i].ID == incidentID {
			checked := time.Now().UTC()
			s.data.Incidents[i].Verification = &Verification{Status: status, CheckedAt: checked, Signals: verifiedSignals}
			if status == "recovered" {
				s.data.Incidents[i].Status = "resolved"
			} else {
				s.data.Incidents[i].Status = "unresolved"
			}
			s.data.Incidents[i].UpdatedAt = checked
			s.appendTimeline(&s.data.Incidents[i], "recovery_verified", "Recovery verification: "+status, "mica", "Mica")
			return s.data.Incidents[i], s.saveLocked()
		}
	}
	return Incident{}, errors.New("incident not found")
}

func (s *Store) AddNote(incidentID, note, actor string) (Incident, error) {
	if note == "" {
		return Incident{}, errors.New("note is required")
	}
	if actor == "" {
		actor = "Engineer"
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.data.Incidents {
		if s.data.Incidents[i].ID == incidentID {
			s.appendTimeline(&s.data.Incidents[i], "note_added", note, "human", actor)
			s.data.Incidents[i].UpdatedAt = time.Now().UTC()
			return s.data.Incidents[i], s.saveLocked()
		}
	}
	return Incident{}, errors.New("incident not found")
}

func (s *Store) appendTimeline(incident *Incident, eventType, summary, actorType, actorName string) {
	event := TimelineEvent{ID: fmt.Sprintf("evt_%d", time.Now().UnixNano()), Type: eventType, Summary: summary, ActorType: actorType, ActorName: actorName, OccurredAt: time.Now().UTC()}
	incident.Timeline = append(incident.Timeline, event)
	if s.timeline != nil {
		s.timeline.Publish(event)
	}
}

func (s *Store) saveLocked() error {
	contents, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return storage.Save(s.database, contents)
}
