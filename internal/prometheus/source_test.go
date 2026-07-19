package prometheus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/mica-dev/mica/internal/incidents"
)

func TestCompareUsesRangeWindowsAndCalculatesDegradation(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/query_range" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Fatalf("authorization = %q", got)
		}
		requests++
		value := "100"
		if requests == 2 {
			value = "150"
		}
		_, _ = w.Write([]byte(`{"status":"success","data":{"result":[{"values":[[1,"` + value + `"],[2,"` + value + `"]]}]}}`))
	}))
	defer server.Close()

	source := Source{BaseURL: server.URL, BearerToken: "test-token", Signals: []SignalSpec{{Name: "latency", Unit: "ms", Query: "demo_latency", TolerancePct: 20}}}
	now := time.Now().UTC()
	result, err := source.Compare(context.Background(), incidents.Service{}, now.Add(-2*time.Hour), now.Add(-time.Hour), now.Add(-10*time.Minute), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("signals = %d", len(result))
	}
	if result[0].BaselineValue != 100 || result[0].IncidentValue != 150 {
		t.Fatalf("comparison = %#v", result[0])
	}
	if result[0].Classification != "degraded" || result[0].PercentDelta != 50 {
		t.Fatalf("classification = %#v", result[0])
	}
	if requests != 2 {
		t.Fatalf("request count = %d", requests)
	}
}

func TestOptionalSignalDoesNotBreakComparison(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Query().Get("query"), "optional") {
			_, _ = w.Write([]byte(`{"status":"success","data":{"result":[]}}`))
			return
		}
		_, _ = w.Write([]byte(`{"status":"success","data":{"result":[{"values":[[1,"10"]]}]}}`))
	}))
	defer server.Close()
	source := Source{BaseURL: server.URL, Signals: []SignalSpec{{Name: "required", Query: "required", TolerancePct: 10}, {Name: "optional", Query: "optional", Optional: true}}}
	now := time.Now()
	result, err := source.Compare(context.Background(), incidents.Service{}, now.Add(-time.Hour), now.Add(-30*time.Minute), now.Add(-10*time.Minute), now)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].Signal != "required" {
		t.Fatalf("result = %#v", result)
	}
}
