package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mica-dev/mica/internal/communications"
	"github.com/mica-dev/mica/internal/events"
	"github.com/mica-dev/mica/internal/incidents"
)

type Handler struct {
	store          *incidents.Store
	events         *events.Bus
	communications *communications.Service
	web            fs.FS
}

func NewHandler(store *incidents.Store) http.Handler {
	return newHandler(store, nil)
}

// NewHandlerWithWeb serves the generated React workspace from the executable.
func NewHandlerWithWeb(store *incidents.Store, web fs.FS) http.Handler {
	return newHandler(store, web)
}

func newHandler(store *incidents.Store, web fs.FS) http.Handler {
	bus := events.New()
	store.SetTimelineSink(bus)
	h := Handler{store: store, events: bus, communications: communications.NewFromEnv(), web: web}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("GET /api/services", h.services)
	mux.HandleFunc("GET /api/services/", h.service)
	mux.HandleFunc("GET /api/incidents", h.incidents)
	mux.HandleFunc("GET /api/communications/destinations", h.destinations)
	mux.HandleFunc("POST /api/incidents/detect", h.detect)
	mux.HandleFunc("GET /api/events", h.streamEvents)
	mux.HandleFunc("GET /api/incidents/", h.incidentRoutes)
	mux.HandleFunc("POST /api/incidents/", h.incidentRoutes)
	mux.HandleFunc("GET /", h.workspace)
	return logging(mux)
}

func (h Handler) health(w http.ResponseWriter, _ *http.Request) {
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}
func (h Handler) services(w http.ResponseWriter, _ *http.Request) {
	respond(w, http.StatusOK, h.store.Services())
}
func (h Handler) incidents(w http.ResponseWriter, _ *http.Request) {
	respond(w, http.StatusOK, h.store.Incidents())
}
func (h Handler) destinations(w http.ResponseWriter, _ *http.Request) {
	respond(w, http.StatusOK, h.communications.Destinations())
}

func (h Handler) service(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/services/")
	service, ok := h.store.Service(id)
	if !ok {
		problem(w, http.StatusNotFound, "service not found")
		return
	}
	respond(w, http.StatusOK, service)
}

type detectInput struct {
	ServiceID     string    `json:"serviceId"`
	BaselineStart time.Time `json:"baselineStart"`
	BaselineEnd   time.Time `json:"baselineEnd"`
	IncidentStart time.Time `json:"incidentStart"`
	IncidentEnd   time.Time `json:"incidentEnd"`
}

func (h Handler) detect(w http.ResponseWriter, r *http.Request) {
	var input detectInput
	if err := decode(r, &input); err != nil {
		problem(w, http.StatusBadRequest, err.Error())
		return
	}
	if input.ServiceID == "" {
		problem(w, http.StatusBadRequest, "serviceId is required")
		return
	}
	incident, err := h.store.Detect(input.ServiceID, input.BaselineStart, input.BaselineEnd, input.IncidentStart, input.IncidentEnd)
	if err != nil {
		problem(w, http.StatusBadRequest, err.Error())
		return
	}
	respond(w, http.StatusCreated, incident)
}

func (h Handler) incidentRoutes(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/incidents/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		problem(w, http.StatusNotFound, "incident ID is required")
		return
	}
	id := parts[0]
	if r.Method == http.MethodGet && len(parts) == 1 {
		incident, ok := h.store.Incident(id)
		if !ok {
			problem(w, http.StatusNotFound, "incident not found")
			return
		}
		respond(w, http.StatusOK, incident)
		return
	}
	if r.Method == http.MethodGet && len(parts) == 2 && parts[1] == "report" {
		incident, ok := h.store.Incident(id)
		if !ok {
			problem(w, http.StatusNotFound, "incident not found")
			return
		}
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		_, _ = w.Write([]byte(incidents.MarkdownReport(incident)))
		return
	}
	if r.Method == http.MethodPost && len(parts) == 4 && parts[1] == "proposals" && parts[3] == "review" {
		var input struct {
			Status   string `json:"status"`
			Reviewer string `json:"reviewer"`
			Note     string `json:"note"`
		}
		if err := decode(r, &input); err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		incident, err := h.store.ReviewProposal(id, parts[2], input.Status, input.Reviewer, input.Note)
		if err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		respond(w, http.StatusOK, incident)
		return
	}
	if r.Method == http.MethodPost && len(parts) == 4 && parts[1] == "updates" && parts[3] == "publish" {
		var input struct {
			ApprovedBy string `json:"approvedBy"`
		}
		if err := decode(r, &input); err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		update, ok := h.store.Update(id, parts[2])
		if !ok {
			problem(w, http.StatusNotFound, "prepared update not found")
			return
		}
		incident, err := h.store.RecordUpdateDelivery(id, update.ID, input.ApprovedBy, h.communications.Publish(update))
		if err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		respond(w, http.StatusOK, incident)
		return
	}
	if r.Method != http.MethodPost || len(parts) != 2 {
		problem(w, http.StatusNotFound, "route not found")
		return
	}
	switch parts[1] {
	case "hypotheses":
		var input incidents.Hypothesis
		if err := decode(r, &input); err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		incident, err := h.store.RecordHypothesis(id, input)
		if err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		respond(w, http.StatusOK, incident)
	case "changes":
		var input incidents.Change
		if err := decode(r, &input); err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		if input.Summary == "" {
			problem(w, http.StatusBadRequest, "summary is required")
			return
		}
		incident, err := h.store.RecordChange(id, input)
		if err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		respond(w, http.StatusOK, incident)
	case "verify":
		var input struct {
			VerificationStart time.Time `json:"verificationStart"`
			VerificationEnd   time.Time `json:"verificationEnd"`
		}
		if r.ContentLength > 0 {
			if err := decode(r, &input); err != nil {
				problem(w, http.StatusBadRequest, err.Error())
				return
			}
		}
		incident, err := h.store.VerifyWindow(id, input.VerificationStart, input.VerificationEnd)
		if err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		respond(w, http.StatusOK, incident)
	case "notes":
		var input struct {
			Note  string `json:"note"`
			Actor string `json:"actor"`
		}
		if err := decode(r, &input); err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		incident, err := h.store.AddNote(id, input.Note, input.Actor)
		if err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		respond(w, http.StatusOK, incident)
	case "proposals":
		var input incidents.ActionProposal
		if err := decode(r, &input); err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		incident, err := h.store.ProposeAction(id, input)
		if err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		respond(w, http.StatusOK, incident)
	case "updates":
		var input struct {
			UpdateType     string   `json:"updateType"`
			Audience       string   `json:"audience"`
			DestinationIDs []string `json:"destinationIds"`
			PreparedBy     string   `json:"preparedBy"`
		}
		if err := decode(r, &input); err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		current, ok := h.store.Incident(id)
		if !ok {
			problem(w, http.StatusNotFound, "incident not found")
			return
		}
		update, err := h.communications.Prepare(current, input.UpdateType, input.Audience, input.DestinationIDs, input.PreparedBy)
		if err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		incident, err := h.store.PrepareUpdate(id, update)
		if err != nil {
			problem(w, http.StatusBadRequest, err.Error())
			return
		}
		respond(w, http.StatusOK, incident)
	default:
		problem(w, http.StatusNotFound, "route not found")
	}
}

func (h Handler) streamEvents(w http.ResponseWriter, r *http.Request) {
	writer, ok := w.(http.Flusher)
	if !ok {
		problem(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	incidentID := r.URL.Query().Get("incidentId")
	channel, unsubscribe := h.events.Subscribe()
	defer unsubscribe()
	for {
		select {
		case <-r.Context().Done():
			return
		case event := <-channel:
			if incidentID != "" {
				incident, ok := h.store.Incident(incidentID)
				if !ok || !hasTimelineEvent(incident, event.ID) {
					continue
				}
			}
			encoded, _ := json.Marshal(event)
			_, _ = fmt.Fprintf(w, "event: incident-update\ndata: %s\n\n", encoded)
			writer.Flush()
		}
	}
}

func hasTimelineEvent(incident incidents.Incident, id string) bool {
	for _, event := range incident.Timeline {
		if event.ID == id {
			return true
		}
	}
	return false
}

func (h Handler) workspace(w http.ResponseWriter, r *http.Request) {
	if h.web != nil {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		file, err := h.web.Open(path)
		if err != nil {
			file, err = h.web.Open("index.html")
		}
		if err == nil {
			defer file.Close()
			if content, ok := file.(io.ReadSeeker); ok {
				http.ServeContent(w, r, path, time.Time{}, content)
				return
			}
		}
	}
	// Compatibility fallback for old direct daemon installs.
	if _, err := os.Stat("/web/index.html"); err == nil {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "/web/index.html")
			return
		}
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			http.ServeFile(w, r, filepath.Join("/web", r.URL.Path))
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(workspaceHTML))
}

func decode(r *http.Request, target any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return errors.New("invalid JSON request body")
	}
	return nil
}
func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func problem(w http.ResponseWriter, status int, detail string) {
	respond(w, status, map[string]string{"error": detail})
}
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}
