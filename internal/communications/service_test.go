package communications

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mica-dev/mica/internal/incidents"
)

func TestProviderPayloadsUseProviderContracts(t *testing.T) {
	for _, test := range []struct {
		provider string
		chatID   string
		want     map[string]string
	}{
		{provider: "slack", want: map[string]string{"text": "update"}},
		{provider: "discord", want: map[string]string{"content": "update"}},
		{provider: "telegram", chatID: "123", want: map[string]string{"chat_id": "123", "text": "update"}},
	} {
		t.Run(test.provider, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var got map[string]string
				if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
					t.Fatal(err)
				}
				if len(got) != len(test.want) {
					t.Fatalf("payload = %#v", got)
				}
				for key, value := range test.want {
					if got[key] != value {
						t.Fatalf("payload[%q] = %q", key, got[key])
					}
				}
				w.WriteHeader(http.StatusNoContent)
			}))
			defer server.Close()
			service := &Service{destinations: map[string]Destination{"target": {ID: "target", Provider: test.provider, endpoint: server.URL, chatID: test.chatID}}, client: server.Client()}
			receipts := service.Publish(incidents.IncidentUpdate{Body: "update", Destinations: []incidents.DeliveryReceipt{{DestinationID: "target", Status: "pending"}}})
			if receipts[0].Status != "delivered" {
				t.Fatalf("receipt = %#v", receipts[0])
			}
		})
	}
}

func TestPrepareRedactsAndRequiresVerifiedRecovery(t *testing.T) {
	service := &Service{destinations: map[string]Destination{"slack": {ID: "slack", Provider: "slack", endpoint: "unused"}}}
	incident := incidents.Incident{ID: "inc_1", ServiceID: "checkout", Evidence: []incidents.SignalComparison{{Signal: "p95 latency", BaselineValue: 120, IncidentValue: 410, Unit: "ms", PercentDelta: 242}}}
	if _, err := service.Prepare(incident, "recovery_verified", "engineering", nil, "agent"); err == nil {
		t.Fatal("expected unverified recovery to fail")
	}
	update, err := service.Prepare(incident, "investigation_update", "engineering", []string{"slack"}, "agent")
	if err != nil {
		t.Fatal(err)
	}
	if update.Status != "" || update.Destinations[0].Status != "pending" || !strings.Contains(update.Body, "p95 latency") {
		t.Fatalf("update = %#v", update)
	}
	if len(update.Redactions) == 0 {
		t.Fatal("expected redactions")
	}
}

func TestPrepareAllowsLocalDraftWithoutDestination(t *testing.T) {
	service := &Service{destinations: map[string]Destination{}}
	update, err := service.Prepare(incidents.Incident{ID: "inc_local", ServiceID: "checkout"}, "investigation_update", "engineering", nil, "Operator")
	if err != nil {
		t.Fatalf("Prepare returned error: %v", err)
	}
	if len(update.Destinations) != 0 {
		t.Fatalf("expected a local-only draft, got %#v", update.Destinations)
	}
	if update.Body == "" || update.PreparedBy != "Operator" {
		t.Fatalf("draft was not populated: %#v", update)
	}
}

func TestPublishUsesConfiguredWebhookWithoutLeakingIt(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s", r.Method)
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			t.Fatal("missing JSON content type")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	service := &Service{destinations: map[string]Destination{"slack": {ID: "slack", Provider: "slack", endpoint: server.URL}}, client: server.Client()}
	receipts := service.Publish(incidents.IncidentUpdate{Body: "safe update", Destinations: []incidents.DeliveryReceipt{{DestinationID: "slack", Provider: "slack", Status: "pending"}}})
	if receipts[0].Status != "delivered" || receipts[0].Response != "accepted" || strings.Contains(receipts[0].Response, server.URL) {
		t.Fatalf("receipt = %#v", receipts[0])
	}
	_ = service.Publish(incidents.IncidentUpdate{Body: "safe update", Destinations: receipts})
	if calls != 1 {
		t.Fatalf("delivered destination was resent %d times", calls)
	}
}
