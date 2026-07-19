package mcpserver

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mica-dev/mica/internal/incidents"
)

func TestToolsListAndIncidentWorkflow(t *testing.T) {
	store, err := incidents.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	input := strings.Join([]string{
		`{"jsonrpc":"2.0","id":1,"method":"tools/list"}`,
		`{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"detect_regressions","arguments":{"serviceId":"checkout"}}}`,
	}, "\n")
	var output bytes.Buffer
	if err := New(store).Serve(strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("responses = %q", output.String())
	}
	var listed struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[0]), &listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Result.Tools) != 11 || listed.Result.Tools[0].Name != "get_service_context" || listed.Result.Tools[1].Name != "record_skill_run" {
		t.Fatalf("tools = %#v", listed.Result.Tools)
	}
	var detected struct {
		Result struct {
			StructuredContent incidents.Incident `json:"structuredContent"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(lines[1]), &detected); err != nil {
		t.Fatal(err)
	}
	if detected.Result.StructuredContent.ID == "" || len(detected.Result.StructuredContent.Evidence) != 3 {
		t.Fatalf("incident = %#v", detected.Result.StructuredContent)
	}
	persisted, ok := store.Incident(detected.Result.StructuredContent.ID)
	if !ok || len(persisted.Timeline) != 2 || persisted.Timeline[1].Type != "capability_invoked" {
		t.Fatalf("timeline = %#v", persisted.Timeline)
	}
}

func TestRejectsInvalidToolArguments(t *testing.T) {
	store, err := incidents.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	input := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"record_change","arguments":{}}}`
	if err := New(store).Serve(strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), `"code":-32602`) {
		t.Fatalf("response = %s", output.String())
	}
}

func TestReadsIncidentTimelineResource(t *testing.T) {
	store, err := incidents.Open(filepath.Join(t.TempDir(), "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	incident, err := store.Detect("checkout", time.Time{}, time.Time{}, time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	input := `{"jsonrpc":"2.0","id":1,"method":"resources/read","params":{"uri":"mica://incidents/` + incident.ID + `/timeline"}}`
	if err := New(store).Serve(strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "regression_detected") {
		t.Fatalf("response = %s", output.String())
	}
}
