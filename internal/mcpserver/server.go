// Package mcpserver provides Mica's stdio MCP surface without exposing raw
// PromQL or write access to production systems.
package mcpserver

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/mica-dev/mica/internal/communications"
	"github.com/mica-dev/mica/internal/incidents"
)

type Server struct{ store *incidents.Store }

func New(store *incidents.Store) *Server { return &Server{store: store} }

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}
type response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}
type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Serve processes newline-delimited JSON-RPC 2.0 messages, the stdio transport
// used by MCP clients. Log output must be directed to stderr by the caller.
func (s *Server) Serve(input io.Reader, output io.Writer) error {
	scanner := bufio.NewScanner(input)
	scanner.Buffer(make([]byte, 4<<10), 1<<20)
	encoder := json.NewEncoder(output)
	for scanner.Scan() {
		var request request
		if err := json.Unmarshal(scanner.Bytes(), &request); err != nil {
			_ = encoder.Encode(response{JSONRPC: "2.0", Error: &rpcError{Code: -32700, Message: "invalid JSON-RPC message"}})
			continue
		}
		if len(request.ID) == 0 {
			continue
		}
		result, rpcErr := s.dispatch(request)
		reply := response{JSONRPC: "2.0", ID: request.ID, Result: result, Error: rpcErr}
		if err := encoder.Encode(reply); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s *Server) dispatch(request request) (any, *rpcError) {
	switch request.Method {
	case "initialize":
		return map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{"tools": map[string]any{}, "resources": map[string]any{}}, "serverInfo": map[string]string{"name": "mica", "version": "0.1.0"}}, nil
	case "notifications/initialized":
		return map[string]any{}, nil
	case "tools/list":
		return map[string]any{"tools": tools()}, nil
	case "resources/list":
		return map[string]any{"resources": []any{map[string]string{"uri": "mica://services", "name": "Mica services", "mimeType": "application/json"}, map[string]string{"uri": "mica://incidents/{incidentId}", "name": "Mica incident", "mimeType": "application/json"}, map[string]string{"uri": "mica://incidents/{incidentId}/timeline", "name": "Mica incident timeline", "mimeType": "application/json"}}}, nil
	case "resources/read":
		return s.readResource(request.Params)
	case "tools/call":
		return s.callTool(request.Params)
	default:
		return nil, &rpcError{Code: -32601, Message: "method not found"}
	}
}

func tools() []map[string]any {
	return []map[string]any{
		{"name": "get_service_context", "description": "Retrieve configured service context, dependencies, signal mappings, sources, and freshness.", "inputSchema": objectSchema(map[string]any{"serviceId": stringProperty("Configured Mica service ID")}, []string{"serviceId"})},
		{"name": "record_skill_run", "description": "Record the reviewed Mica skill and version guiding this incident workflow in the shared timeline.", "inputSchema": objectSchema(map[string]any{"incidentId": stringProperty("Mica incident ID"), "skillName": stringProperty("Skill name"), "version": stringProperty("Skill version")}, []string{"incidentId", "skillName", "version"})},
		{"name": "detect_regressions", "description": "Compare a service's current window to a saved healthy baseline and create a persisted incident with evidence IDs.", "inputSchema": objectSchema(map[string]any{"serviceId": stringProperty("Configured Mica service ID"), "baselineStart": stringProperty("RFC3339 timestamp"), "baselineEnd": stringProperty("RFC3339 timestamp"), "incidentStart": stringProperty("RFC3339 timestamp"), "incidentEnd": stringProperty("RFC3339 timestamp")}, []string{"serviceId"})},
		{"name": "inspect_service", "description": "Retrieve persisted incident evidence for a service without constructing raw PromQL.", "inputSchema": objectSchema(map[string]any{"incidentId": stringProperty("Mica incident ID")}, []string{"incidentId"})},
		{"name": "find_correlations", "description": "Return deterministic signal associations from saved incident evidence. Association is not proof of causality.", "inputSchema": objectSchema(map[string]any{"incidentId": stringProperty("Mica incident ID"), "evidenceId": stringProperty("Primary evidence ID")}, []string{"incidentId", "evidenceId"})},
		{"name": "record_hypothesis", "description": "Record an evidence-linked investigation hypothesis in Mica. This only mutates Mica incident state.", "inputSchema": objectSchema(map[string]any{"incidentId": stringProperty("Mica incident ID"), "summary": stringProperty("Evidence-backed hypothesis"), "confidence": stringProperty("low, medium, or high"), "evidenceIds": map[string]any{"type": "array", "items": map[string]string{"type": "string"}}, "alternatives": map[string]any{"type": "array", "items": map[string]string{"type": "string"}}, "nextStep": stringProperty("Safe next investigation step")}, []string{"incidentId", "summary", "evidenceIds", "alternatives", "nextStep"})},
		{"name": "record_change", "description": "Record a completed code change and tests. It cannot deploy or change production infrastructure.", "inputSchema": objectSchema(map[string]any{"incidentId": stringProperty("Mica incident ID"), "summary": stringProperty("Concise change summary"), "files": map[string]any{"type": "array", "items": map[string]string{"type": "string"}}, "tests": map[string]any{"type": "array", "items": map[string]string{"type": "string"}}}, []string{"incidentId", "summary"})},
		{"name": "propose_action", "description": "Create a structured, proposal-only operational next step. This does not execute any external operation; humans may only review, reject, or defer it in the MVP.", "inputSchema": objectSchema(map[string]any{"incidentId": stringProperty("Mica incident ID"), "actionType": stringProperty("Bounded action type"), "summary": stringProperty("Action summary"), "parameters": map[string]any{"type": "object", "additionalProperties": map[string]string{"type": "string"}}, "expectedEffect": stringProperty("Expected result"), "riskStatement": stringProperty("Known risk and scope"), "verificationPlan": stringProperty("How success will be checked"), "evidenceIds": map[string]any{"type": "array", "items": map[string]string{"type": "string"}}, "riskLevel": stringProperty("Mica permission level; defaults to propose")}, []string{"incidentId", "actionType", "summary", "expectedEffect", "riskStatement", "verificationPlan", "evidenceIds"})},
		{"name": "prepare_incident_update", "description": "Prepare a grounded, redacted update for configured destinations. This never publishes externally and must be followed by explicit human approval in Mica.", "inputSchema": objectSchema(map[string]any{"incidentId": stringProperty("Mica incident ID"), "updateType": stringProperty("incident_opened, investigation_update, approval_needed, mitigation_applied, recovery_verified, or handoff"), "audience": stringProperty("Intended audience"), "destinationIds": map[string]any{"type": "array", "items": map[string]string{"type": "string"}}}, []string{"incidentId", "updateType", "audience"})},
		{"name": "record_audit_findings", "description": "Persist structured evidence-backed readiness, security, or release-risk findings. High and critical findings require evidence references and every finding requires a verification method.", "inputSchema": objectSchema(map[string]any{"incidentId": stringProperty("Mica incident ID"), "findings": map[string]any{"type": "array", "items": map[string]any{"type": "object"}}}, []string{"incidentId", "findings"})},
		{"name": "verify_recovery", "description": "Verify fresh telemetry against the incident's original baseline, required signals, and tolerances. It never redefines success criteria.", "inputSchema": objectSchema(map[string]any{"incidentId": stringProperty("Mica incident ID"), "verificationStart": stringProperty("RFC3339 timestamp"), "verificationEnd": stringProperty("RFC3339 timestamp")}, []string{"incidentId"})},
	}
}

func objectSchema(properties map[string]any, required []string) map[string]any {
	return map[string]any{"type": "object", "properties": properties, "required": required, "additionalProperties": false}
}
func stringProperty(description string) map[string]string {
	return map[string]string{"type": "string", "description": description}
}

func (s *Server) readResource(raw json.RawMessage) (any, *rpcError) {
	var input struct {
		URI string `json:"uri"`
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, invalidParams()
	}
	if input.URI == "mica://services" {
		return resource(input.URI, s.store.Services()), nil
	}
	if strings.HasPrefix(input.URI, "mica://incidents/") {
		path := strings.TrimPrefix(input.URI, "mica://incidents/")
		parts := strings.Split(path, "/")
		id := parts[0]
		incident, ok := s.store.Incident(id)
		if !ok {
			return nil, &rpcError{Code: -32004, Message: "incident not found"}
		}
		if len(parts) == 2 && parts[1] == "timeline" {
			return resource(input.URI, incident.Timeline), nil
		}
		if len(parts) != 1 {
			return nil, &rpcError{Code: -32004, Message: "resource not found"}
		}
		return resource(input.URI, incident), nil
	}
	return nil, &rpcError{Code: -32004, Message: "resource not found"}
}

func (s *Server) callTool(raw json.RawMessage) (any, *rpcError) {
	var input struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return nil, invalidParams()
	}
	var output any
	var err error
	var invocationIncidentID string
	switch input.Name {
	case "record_skill_run":
		var args struct {
			IncidentID string `json:"incidentId"`
			SkillName  string `json:"skillName"`
			Version    string `json:"version"`
		}
		if json.Unmarshal(input.Arguments, &args) != nil || args.IncidentID == "" || args.SkillName == "" || args.Version == "" {
			return nil, invalidParams()
		}
		err = s.store.RecordSkillInvocation(args.IncidentID, args.SkillName, args.Version)
		output = map[string]string{"incidentId": args.IncidentID, "skillName": args.SkillName, "version": args.Version}
		invocationIncidentID = args.IncidentID
	case "get_service_context":
		var args struct {
			ServiceID string `json:"serviceId"`
		}
		if json.Unmarshal(input.Arguments, &args) != nil || args.ServiceID == "" {
			return nil, invalidParams()
		}
		value, ok := s.store.Service(args.ServiceID)
		if !ok {
			return toolError("service not found")
		}
		output = value
	case "detect_regressions":
		var args struct {
			ServiceID                                              string `json:"serviceId"`
			BaselineStart, BaselineEnd, IncidentStart, IncidentEnd string
		}
		if json.Unmarshal(input.Arguments, &args) != nil || args.ServiceID == "" {
			return nil, invalidParams()
		}
		output, err = s.store.Detect(args.ServiceID, parseTime(args.BaselineStart), parseTime(args.BaselineEnd), parseTime(args.IncidentStart), parseTime(args.IncidentEnd))
		if value, ok := output.(incidents.Incident); ok {
			invocationIncidentID = value.ID
		}
	case "inspect_service":
		var args struct {
			IncidentID string `json:"incidentId"`
		}
		if json.Unmarshal(input.Arguments, &args) != nil || args.IncidentID == "" {
			return nil, invalidParams()
		}
		value, ok := s.store.Incident(args.IncidentID)
		if !ok {
			return toolError("incident not found")
		}
		output = map[string]any{"incidentId": value.ID, "status": value.Status, "evidence": value.Evidence}
		invocationIncidentID = args.IncidentID
	case "find_correlations":
		var args struct {
			IncidentID string `json:"incidentId"`
			EvidenceID string `json:"evidenceId"`
		}
		if json.Unmarshal(input.Arguments, &args) != nil || args.IncidentID == "" || args.EvidenceID == "" {
			return nil, invalidParams()
		}
		value, ok := s.store.Incident(args.IncidentID)
		if !ok {
			return toolError("incident not found")
		}
		primary := incidents.SignalComparison{}
		for _, evidence := range value.Evidence {
			if evidence.ID == args.EvidenceID {
				primary = evidence
				break
			}
		}
		if primary.ID == "" {
			return toolError("evidence ID does not belong to incident")
		}
		associations := make([]map[string]any, 0, len(value.Evidence)-1)
		for _, evidence := range value.Evidence {
			if evidence.ID == primary.ID {
				continue
			}
			associations = append(associations, map[string]any{"evidenceId": evidence.ID, "signal": evidence.Signal, "association": "co-regressed in the same incident window", "score": associationScore(primary.PercentDelta, evidence.PercentDelta), "causality": "unproven"})
		}
		output = map[string]any{"incidentId": value.ID, "primaryEvidenceId": primary.ID, "associations": associations}
		invocationIncidentID = args.IncidentID
	case "record_hypothesis":
		var args struct {
			IncidentID   string   `json:"incidentId"`
			Summary      string   `json:"summary"`
			Confidence   string   `json:"confidence"`
			EvidenceIDs  []string `json:"evidenceIds"`
			Alternatives []string `json:"alternatives"`
			NextStep     string   `json:"nextStep"`
		}
		if json.Unmarshal(input.Arguments, &args) != nil || args.IncidentID == "" || args.Summary == "" {
			return nil, invalidParams()
		}
		output, err = s.store.RecordHypothesis(args.IncidentID, incidents.Hypothesis{Summary: args.Summary, Confidence: args.Confidence, EvidenceIDs: args.EvidenceIDs, Alternatives: args.Alternatives, NextStep: args.NextStep})
		invocationIncidentID = args.IncidentID
	case "record_change":
		var args struct {
			IncidentID string   `json:"incidentId"`
			Summary    string   `json:"summary"`
			Files      []string `json:"files"`
			Tests      []string `json:"tests"`
		}
		if json.Unmarshal(input.Arguments, &args) != nil || args.IncidentID == "" || args.Summary == "" {
			return nil, invalidParams()
		}
		output, err = s.store.RecordChange(args.IncidentID, incidents.Change{Summary: args.Summary, Files: args.Files, Tests: args.Tests})
		invocationIncidentID = args.IncidentID
	case "propose_action":
		var args struct {
			IncidentID       string            `json:"incidentId"`
			ActionType       string            `json:"actionType"`
			Summary          string            `json:"summary"`
			Parameters       map[string]string `json:"parameters"`
			ExpectedEffect   string            `json:"expectedEffect"`
			RiskStatement    string            `json:"riskStatement"`
			VerificationPlan string            `json:"verificationPlan"`
			EvidenceIDs      []string          `json:"evidenceIds"`
			RiskLevel        string            `json:"riskLevel"`
		}
		if json.Unmarshal(input.Arguments, &args) != nil || args.IncidentID == "" {
			return nil, invalidParams()
		}
		output, err = s.store.ProposeAction(args.IncidentID, incidents.ActionProposal{ActionType: args.ActionType, Summary: args.Summary, Parameters: args.Parameters, ExpectedEffect: args.ExpectedEffect, RiskStatement: args.RiskStatement, VerificationPlan: args.VerificationPlan, EvidenceIDs: args.EvidenceIDs, RiskLevel: args.RiskLevel, ProposedBy: "Codex via MCP"})
		invocationIncidentID = args.IncidentID
	case "prepare_incident_update":
		var args struct {
			IncidentID     string   `json:"incidentId"`
			UpdateType     string   `json:"updateType"`
			Audience       string   `json:"audience"`
			DestinationIDs []string `json:"destinationIds"`
		}
		if json.Unmarshal(input.Arguments, &args) != nil || args.IncidentID == "" {
			return nil, invalidParams()
		}
		incident, ok := s.store.Incident(args.IncidentID)
		if !ok {
			return toolError("incident not found")
		}
		update, prepareErr := communications.NewFromEnv().Prepare(incident, args.UpdateType, args.Audience, args.DestinationIDs, "Codex via MCP")
		if prepareErr != nil {
			return toolError(prepareErr.Error())
		}
		output, err = s.store.PrepareUpdate(args.IncidentID, update)
		invocationIncidentID = args.IncidentID
	case "record_audit_findings":
		var args struct {
			IncidentID string                   `json:"incidentId"`
			Findings   []incidents.AuditFinding `json:"findings"`
		}
		if json.Unmarshal(input.Arguments, &args) != nil || args.IncidentID == "" {
			return nil, invalidParams()
		}
		output, err = s.store.RecordAuditFindings(args.IncidentID, args.Findings, "Codex via MCP")
		invocationIncidentID = args.IncidentID
	case "verify_recovery":
		var args struct {
			IncidentID                         string `json:"incidentId"`
			VerificationStart, VerificationEnd string
		}
		if json.Unmarshal(input.Arguments, &args) != nil || args.IncidentID == "" {
			return nil, invalidParams()
		}
		output, err = s.store.VerifyWindow(args.IncidentID, parseTime(args.VerificationStart), parseTime(args.VerificationEnd))
		invocationIncidentID = args.IncidentID
	default:
		return nil, &rpcError{Code: -32601, Message: "tool not found"}
	}
	if err != nil {
		return toolError(err.Error())
	}
	if invocationIncidentID != "" {
		if err := s.store.RecordCapabilityInvocation(invocationIncidentID, input.Name); err != nil {
			return toolError(err.Error())
		}
	}
	return toolResult(output), nil
}

func resource(uri string, value any) map[string]any {
	encoded, _ := json.Marshal(value)
	return map[string]any{"contents": []map[string]string{{"uri": uri, "mimeType": "application/json", "text": string(encoded)}}}
}
func toolResult(value any) map[string]any {
	encoded, _ := json.Marshal(value)
	return map[string]any{"content": []map[string]string{{"type": "text", "text": string(encoded)}}, "structuredContent": value}
}
func toolError(message string) (any, *rpcError) {
	return nil, &rpcError{Code: -32000, Message: message}
}
func invalidParams() *rpcError         { return &rpcError{Code: -32602, Message: "invalid parameters"} }
func parseTime(value string) time.Time { parsed, _ := time.Parse(time.RFC3339, value); return parsed }

func associationScore(first, second float64) float64 {
	if first == 0 || second == 0 || (first < 0) != (second < 0) {
		return 0
	}
	return 0.9
}
