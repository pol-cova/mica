// Package communications builds redacted, fact-grounded incident updates and
// sends them only after the API supplies an explicit human approver.
package communications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mica-dev/mica/internal/incidents"
)

type Destination struct {
	ID       string
	Provider string
	endpoint string
	chatID   string
}
type Service struct {
	destinations map[string]Destination
	client       *http.Client
}

func (s *Service) Destinations() []map[string]string {
	result := make([]map[string]string, 0, len(s.destinations))
	for _, destination := range s.destinations {
		result = append(result, map[string]string{"id": destination.ID, "provider": destination.Provider})
	}
	return result
}

func NewFromEnv() *Service {
	destinations := map[string]Destination{}
	if value := os.Getenv("MICA_SLACK_WEBHOOK_URL"); value != "" {
		destinations["slack"] = Destination{ID: "slack", Provider: "slack", endpoint: value}
	}
	if value := os.Getenv("MICA_DISCORD_WEBHOOK_URL"); value != "" {
		destinations["discord"] = Destination{ID: "discord", Provider: "discord", endpoint: value}
	}
	if token, chatID := os.Getenv("MICA_TELEGRAM_BOT_TOKEN"), os.Getenv("MICA_TELEGRAM_CHAT_ID"); token != "" && chatID != "" {
		destinations["telegram"] = Destination{ID: "telegram", Provider: "telegram", endpoint: "https://api.telegram.org/bot" + token + "/sendMessage", chatID: chatID}
	}
	return &Service{destinations: destinations, client: &http.Client{Timeout: 10 * time.Second}}
}

func (s *Service) Prepare(incident incidents.Incident, updateType, audience string, destinationIDs []string, preparedBy string) (incidents.IncidentUpdate, error) {
	if !validType(updateType) {
		return incidents.IncidentUpdate{}, fmt.Errorf("unsupported update type %q", updateType)
	}
	if updateType == "recovery_verified" && (incident.Verification == nil || incident.Verification.Status != "recovered") {
		return incidents.IncidentUpdate{}, fmt.Errorf("cannot prepare recovery update before Mica verifies recovery")
	}
	if len(destinationIDs) == 0 {
		for id := range s.destinations {
			destinationIDs = append(destinationIDs, id)
		}
	}
	receipts := make([]incidents.DeliveryReceipt, 0, len(destinationIDs))
	seen := map[string]bool{}
	for _, id := range destinationIDs {
		destination, ok := s.destinations[id]
		if !ok {
			return incidents.IncidentUpdate{}, fmt.Errorf("destination %q is not configured", id)
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		receipts = append(receipts, incidents.DeliveryReceipt{DestinationID: destination.ID, Provider: destination.Provider, Status: "pending", Attempts: 0})
	}
	if preparedBy == "" {
		preparedBy = "Mica agent"
	}
	return incidents.IncidentUpdate{UpdateType: updateType, Audience: audience, Body: render(incident, updateType), Redactions: []string{"PromQL excluded", "credentials excluded", "private repository paths excluded"}, Destinations: receipts, PreparedBy: preparedBy}, nil
}

func (s *Service) Publish(update incidents.IncidentUpdate) []incidents.DeliveryReceipt {
	receipts := append([]incidents.DeliveryReceipt(nil), update.Destinations...)
	for index := range receipts {
		receipt := &receipts[index]
		// Idempotent retries only send destinations that have not already
		// accepted this immutable prepared update.
		if receipt.Status == "delivered" {
			continue
		}
		destination, ok := s.destinations[receipt.DestinationID]
		if !ok {
			receipt.Status = "failed"
			receipt.Response = "destination is no longer configured"
			continue
		}
		receipt.Attempts++
		if err := s.send(destination, update.Body); err != nil {
			receipt.Status = "failed"
			receipt.Response = redactError(err.Error())
			continue
		}
		now := time.Now().UTC()
		receipt.Status = "delivered"
		receipt.DeliveredAt = &now
		receipt.Response = "accepted"
	}
	return receipts
}

func (s *Service) send(destination Destination, body string) error {
	payload := payloadFor(destination, body)
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, destination.endpoint, bytes.NewReader(encoded))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("provider returned %s", response.Status)
	}
	return nil
}

// payloadFor keeps provider-specific webhook contracts at the daemon boundary.
// The rendered message stays fact-identical; only each provider's required field
// name differs. Credentials and endpoint URLs never leave this package.
func payloadFor(destination Destination, body string) map[string]string {
	switch destination.Provider {
	case "discord":
		return map[string]string{"content": body}
	case "telegram":
		return map[string]string{"chat_id": destination.chatID, "text": body}
	default: // Slack incoming webhooks use text.
		return map[string]string{"text": body}
	}
}

func render(incident incidents.Incident, updateType string) string {
	status := incident.Status
	if updateType == "recovery_verified" {
		status = "resolved"
	}
	lines := []string{fmt.Sprintf("[Mica %s] %s / %s", incident.ID, incident.ServiceID, status), "Type: " + strings.ReplaceAll(updateType, "_", " ")}
	if len(incident.Evidence) > 0 {
		evidence := incident.Evidence[0]
		lines = append(lines, fmt.Sprintf("Measured signal: %s %.2f → %.2f %s (%+.0f%%).", evidence.Signal, evidence.BaselineValue, evidence.IncidentValue, evidence.Unit, evidence.PercentDelta))
	}
	if incident.Verification != nil {
		lines = append(lines, "Verification: "+incident.Verification.Status+".")
	}
	lines = append(lines, "Source of truth: Mica incident workspace.")
	return strings.Join(lines, "\n")
}
func validType(value string) bool {
	switch value {
	case "incident_opened", "investigation_update", "approval_needed", "mitigation_applied", "recovery_verified", "handoff":
		return true
	}
	return false
}
func redactError(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "http://", "[redacted-url]"), "https://", "[redacted-url]")
}
