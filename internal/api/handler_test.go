package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mica-dev/mica/internal/incidents"
)

func TestIncidentHTTPWorkflow(t *testing.T) {
	store, err := incidents.Open(filepath.Join(t.TempDir(), "mica.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	handler := NewHandler(store)

	detected := callJSON(t, handler, http.MethodPost, "/api/incidents/detect", `{"serviceId":"checkout"}`)
	if detected.Code != http.StatusCreated {
		t.Fatalf("detect = %d: %s", detected.Code, detected.Body.String())
	}
	var incident incidents.Incident
	if err := json.NewDecoder(detected.Body).Decode(&incident); err != nil {
		t.Fatal(err)
	}
	if len(incident.Evidence) != 3 {
		t.Fatalf("evidence = %#v", incident.Evidence)
	}

	badHypothesis := callJSON(t, handler, http.MethodPost, "/api/incidents/"+incident.ID+"/hypotheses", `{"summary":"unsupported","evidenceIds":["missing"]}`)
	if badHypothesis.Code != http.StatusBadRequest {
		t.Fatalf("foreign evidence status = %d", badHypothesis.Code)
	}

	proposalBody := `{"actionType":"replay_traffic","summary":"Replay bounded demo traffic","expectedEffect":"Fresh samples","riskStatement":"Demo only","verificationPlan":"Use saved recovery criteria","evidenceIds":["` + incident.Evidence[0].ID + `"]}`
	proposed := callJSON(t, handler, http.MethodPost, "/api/incidents/"+incident.ID+"/proposals", proposalBody)
	if proposed.Code != http.StatusOK {
		t.Fatalf("proposal = %d: %s", proposed.Code, proposed.Body.String())
	}
	if err := json.NewDecoder(proposed.Body).Decode(&incident); err != nil {
		t.Fatal(err)
	}
	if len(incident.Proposals) != 1 || incident.Proposals[0].Status != "pending_review" {
		t.Fatalf("proposals = %#v", incident.Proposals)
	}

	reviewed := callJSON(t, handler, http.MethodPost, "/api/incidents/"+incident.ID+"/proposals/"+incident.Proposals[0].ID+"/review", `{"status":"reviewed","reviewer":"Engineer"}`)
	if reviewed.Code != http.StatusOK {
		t.Fatalf("review = %d", reviewed.Code)
	}

	report := httptest.NewRecorder()
	handler.ServeHTTP(report, httptest.NewRequest(http.MethodGet, "/api/incidents/"+incident.ID+"/report", nil))
	if report.Code != http.StatusOK || !strings.Contains(report.Body.String(), "## Measured evidence") || !strings.Contains(report.Body.String(), "Replay bounded demo traffic") {
		t.Fatalf("report = %d %s", report.Code, report.Body.String())
	}
}

func callJSON(t *testing.T, handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}
