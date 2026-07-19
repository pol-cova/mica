package incidents

import (
	"context"
	"time"
)

// MetricsSource is the narrow port used by incidents to compare operational
// signals. Adapters may query Prometheus or provide deterministic demo data.
type MetricsSource interface {
	Compare(context.Context, Service, time.Time, time.Time, time.Time, time.Time) ([]SignalComparison, error)
}

type Service struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Environment      string            `json:"environment"`
	Repository       string            `json:"repository"`
	Owners           []string          `json:"owners"`
	Dependencies     []string          `json:"dependencies"`
	Signals          map[string]string `json:"signals"`
	Runtime          string            `json:"runtime,omitempty"`
	Framework        string            `json:"framework,omitempty"`
	SLORefs          []string          `json:"sloRefs,omitempty"`
	Runbooks         []string          `json:"runbooks,omitempty"`
	RecentChanges    []string          `json:"recentChanges,omitempty"`
	OperationalNotes []string          `json:"operationalNotes,omitempty"`
	Source           SourceMetadata    `json:"source"`
}

type SourceMetadata struct {
	Kind        string    `json:"kind"`
	RefreshedAt time.Time `json:"refreshedAt"`
}

type SignalComparison struct {
	ID             string  `json:"id"`
	Signal         string  `json:"signal"`
	Unit           string  `json:"unit"`
	Query          string  `json:"query"`
	BaselineValue  float64 `json:"baselineValue"`
	IncidentValue  float64 `json:"incidentValue"`
	CurrentValue   float64 `json:"currentValue,omitempty"`
	AbsoluteDelta  float64 `json:"absoluteDelta"`
	PercentDelta   float64 `json:"percentDelta"`
	Classification string  `json:"classification"`
	TolerancePct   float64 `json:"tolerancePct"`
}

type Incident struct {
	ID            string             `json:"id"`
	ServiceID     string             `json:"serviceId"`
	Status        string             `json:"status"`
	CreatedAt     time.Time          `json:"createdAt"`
	UpdatedAt     time.Time          `json:"updatedAt"`
	BaselineStart time.Time          `json:"baselineStart"`
	BaselineEnd   time.Time          `json:"baselineEnd"`
	IncidentStart time.Time          `json:"incidentStart"`
	IncidentEnd   time.Time          `json:"incidentEnd"`
	Evidence      []SignalComparison `json:"evidence"`
	Hypotheses    []Hypothesis       `json:"hypotheses"`
	Changes       []Change           `json:"changes"`
	Proposals     []ActionProposal   `json:"proposals"`
	Updates       []IncidentUpdate   `json:"updates"`
	AuditFindings []AuditFinding     `json:"auditFindings"`
	Verification  *Verification      `json:"verification,omitempty"`
	Timeline      []TimelineEvent    `json:"timeline"`
}

type TimelineEvent struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Summary    string    `json:"summary"`
	ActorType  string    `json:"actorType"`
	ActorName  string    `json:"actorName"`
	OccurredAt time.Time `json:"occurredAt"`
}

type TimelineSink interface{ Publish(TimelineEvent) }

type Hypothesis struct {
	ID           string    `json:"id"`
	Summary      string    `json:"summary"`
	Confidence   string    `json:"confidence"`
	EvidenceIDs  []string  `json:"evidenceIds"`
	Alternatives []string  `json:"alternatives"`
	NextStep     string    `json:"nextStep"`
	CreatedAt    time.Time `json:"createdAt"`
}

type Change struct {
	ID        string    `json:"id"`
	Summary   string    `json:"summary"`
	Files     []string  `json:"files"`
	Tests     []string  `json:"tests"`
	CreatedAt time.Time `json:"createdAt"`
}

// ActionProposal describes a bounded operational next step. It is deliberately
// proposal-only: review never authorizes external execution in the MVP.
type ActionProposal struct {
	ID               string            `json:"id"`
	ActionType       string            `json:"actionType"`
	Summary          string            `json:"summary"`
	Parameters       map[string]string `json:"parameters"`
	ExpectedEffect   string            `json:"expectedEffect"`
	RiskStatement    string            `json:"riskStatement"`
	VerificationPlan string            `json:"verificationPlan"`
	EvidenceIDs      []string          `json:"evidenceIds"`
	RiskLevel        string            `json:"riskLevel"`
	Status           string            `json:"status"`
	ProposedBy       string            `json:"proposedBy"`
	ReviewedBy       string            `json:"reviewedBy,omitempty"`
	ReviewNote       string            `json:"reviewNote,omitempty"`
	CreatedAt        time.Time         `json:"createdAt"`
	ReviewedAt       *time.Time        `json:"reviewedAt,omitempty"`
}

type IncidentUpdate struct {
	ID             string            `json:"id"`
	UpdateType     string            `json:"updateType"`
	Audience       string            `json:"audience"`
	Body           string            `json:"body"`
	Redactions     []string          `json:"redactions"`
	Destinations   []DeliveryReceipt `json:"destinations"`
	Status         string            `json:"status"`
	PreparedBy     string            `json:"preparedBy"`
	ApprovedBy     string            `json:"approvedBy,omitempty"`
	IdempotencyKey string            `json:"idempotencyKey"`
	CreatedAt      time.Time         `json:"createdAt"`
	ApprovedAt     *time.Time        `json:"approvedAt,omitempty"`
}

type DeliveryReceipt struct {
	DestinationID string     `json:"destinationId"`
	Provider      string     `json:"provider"`
	Status        string     `json:"status"`
	Attempts      int        `json:"attempts"`
	Response      string     `json:"response,omitempty"`
	DeliveredAt   *time.Time `json:"deliveredAt,omitempty"`
}

// AuditFinding is the common structured result for readiness, security, and
// release-risk skills. EvidenceRefs may point to Mica evidence, policy IDs, or
// authorized repository paths; a high-severity finding cannot be evidence-free.
type AuditFinding struct {
	ID                 string    `json:"id"`
	Category           string    `json:"category"`
	Severity           string    `json:"severity"`
	Confidence         string    `json:"confidence"`
	AffectedAsset      string    `json:"affectedAsset"`
	EvidenceRefs       []string  `json:"evidenceRefs"`
	Assessment         string    `json:"assessment"`
	Risk               string    `json:"risk"`
	Recommendation     string    `json:"recommendation"`
	Owner              string    `json:"owner,omitempty"`
	VerificationMethod string    `json:"verificationMethod"`
	Status             string    `json:"status"`
	CreatedAt          time.Time `json:"createdAt"`
}

type Verification struct {
	Status    string             `json:"status"`
	CheckedAt time.Time          `json:"checkedAt"`
	Signals   []SignalComparison `json:"signals"`
}
