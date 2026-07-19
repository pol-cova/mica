package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type counters struct {
	requestsOK  atomic.Uint64
	requests5xx atomic.Uint64
	queries     atomic.Uint64
	buckets     [5]atomic.Uint64
}

var regression atomic.Bool
var calls atomic.Uint64
var metrics counters
var writeMu sync.Mutex

func main() {
	http.HandleFunc("GET /checkout", checkout)
	http.HandleFunc("POST /demo/regression", setRegression)
	http.HandleFunc("GET /demo/status", status)
	http.HandleFunc("GET /metrics", serveMetrics)
	log.Println("checkout demo service listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func checkout(w http.ResponseWriter, r *http.Request) {
	items, _ := strconv.Atoi(r.URL.Query().Get("items"))
	if items < 1 {
		items = 3
	}
	if items > 25 {
		items = 25
	}
	started := time.Now()
	if regression.Load() {
		// The intentional N+1: one query per checkout item plus the order query.
		metrics.queries.Add(uint64(items + 1))
		time.Sleep(time.Duration(items*14) * time.Millisecond)
		if calls.Add(1)%12 == 0 {
			metrics.requests5xx.Add(1)
			http.Error(w, "simulated checkout failure", http.StatusInternalServerError)
			observe(time.Since(started))
			return
		}
	} else {
		// The fixed path uses one batch lookup and one order query.
		metrics.queries.Add(2)
		time.Sleep(18 * time.Millisecond)
	}
	metrics.requestsOK.Add(1)
	observe(time.Since(started))
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func observe(duration time.Duration) {
	seconds := duration.Seconds()
	limits := []float64{0.05, 0.1, 0.25, 0.5, 1}
	for index, limit := range limits {
		if seconds <= limit {
			metrics.buckets[index].Add(1)
		}
	}
}

func setRegression(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "expected JSON body", http.StatusBadRequest)
		return
	}
	regression.Store(input.Enabled)
	writeMu.Lock()
	calls.Store(0)
	writeMu.Unlock()
	status(w, r)
}

func status(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]bool{"regression": regression.Load()})
}

func serveMetrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	ok, failures := metrics.requestsOK.Load(), metrics.requests5xx.Load()
	writeMu.Lock()
	defer writeMu.Unlock()
	fmt.Fprintln(w, "# HELP checkout_http_requests_total Checkout HTTP requests by response code")
	fmt.Fprintln(w, "# TYPE checkout_http_requests_total counter")
	fmt.Fprintf(w, "checkout_http_requests_total{code=\"200\"} %d\n", ok)
	fmt.Fprintf(w, "checkout_http_requests_total{code=\"500\"} %d\n", failures)
	fmt.Fprintln(w, "# HELP checkout_db_queries_total Database queries performed by checkout")
	fmt.Fprintln(w, "# TYPE checkout_db_queries_total counter")
	fmt.Fprintf(w, "checkout_db_queries_total %d\n", metrics.queries.Load())
	fmt.Fprintln(w, "# HELP checkout_http_duration_seconds Checkout response latency histogram")
	fmt.Fprintln(w, "# TYPE checkout_http_duration_seconds histogram")
	limits := []string{"0.05", "0.1", "0.25", "0.5", "1"}
	for index, limit := range limits {
		fmt.Fprintf(w, "checkout_http_duration_seconds_bucket{le=\"%s\"} %d\n", limit, metrics.buckets[index].Load())
	}
	fmt.Fprintf(w, "checkout_http_duration_seconds_bucket{le=\"+Inf\"} %d\n", ok+failures)
	fmt.Fprintf(w, "checkout_http_duration_seconds_count %d\n", ok+failures)
}
