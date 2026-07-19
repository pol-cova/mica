// Package prometheus adapts the read-only Prometheus HTTP API to Mica's
// incident metrics port.
package prometheus

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mica-dev/mica/internal/incidents"
)

type SignalSpec struct {
	Name         string
	Unit         string
	Query        string
	TolerancePct float64
	Optional     bool
}

type Source struct {
	BaseURL     string
	BearerToken string
	BasicUser   string
	BasicPass   string
	Client      *http.Client
	Signals     []SignalSpec
	Step        time.Duration
}

func DefaultCheckoutSignals() []SignalSpec {
	return CheckoutSignals("5m")
}

// CheckoutSignals builds the demo mappings with a configurable range-vector
// duration. Production defaults to five minutes; the bundled short demo can
// opt into a smaller validated interval without changing the signal model.
func CheckoutSignals(rateWindow string) []SignalSpec {
	if _, err := time.ParseDuration(rateWindow); err != nil || rateWindow == "" {
		rateWindow = "5m"
	}
	return []SignalSpec{
		{Name: "p95 latency", Unit: "ms", Query: fmt.Sprintf("histogram_quantile(0.95, sum(rate(checkout_http_duration_seconds_bucket[%s])) by (le)) * 1000", rateWindow), TolerancePct: 20},
		{Name: "error rate", Unit: "%", Query: fmt.Sprintf("sum(rate(checkout_http_requests_total{code=~\"5..\"}[%s])) / sum(rate(checkout_http_requests_total[%s])) * 100", rateWindow, rateWindow), TolerancePct: 25},
		{Name: "database queries per request", Unit: "queries/request", Query: fmt.Sprintf("sum(rate(checkout_db_queries_total[%s])) / sum(rate(checkout_http_requests_total[%s]))", rateWindow, rateWindow), TolerancePct: 20, Optional: true},
	}
}

func (s Source) Compare(ctx context.Context, _ incidents.Service, baselineStart, baselineEnd, incidentStart, incidentEnd time.Time) ([]incidents.SignalComparison, error) {
	if s.BaseURL == "" {
		return nil, fmt.Errorf("Prometheus base URL is required")
	}
	signals := s.Signals
	if len(signals) == 0 {
		signals = DefaultCheckoutSignals()
	}
	result := make([]incidents.SignalComparison, 0, len(signals))
	for index, signal := range signals {
		baseline, err := s.averageRange(ctx, signal.Query, baselineStart, baselineEnd)
		if err != nil {
			if signal.Optional {
				continue
			}
			return nil, fmt.Errorf("%s baseline: %w", signal.Name, err)
		}
		current, err := s.averageRange(ctx, signal.Query, incidentStart, incidentEnd)
		if err != nil {
			if signal.Optional {
				continue
			}
			return nil, fmt.Errorf("%s incident: %w", signal.Name, err)
		}
		comparison := incidents.SignalComparison{ID: fmt.Sprintf("ev_%d", index+1), Signal: signal.Name, Unit: signal.Unit, Query: signal.Query, BaselineValue: baseline, IncidentValue: current, AbsoluteDelta: current - baseline, TolerancePct: signal.TolerancePct}
		if baseline == 0 {
			comparison.Classification = "insufficient_data"
		} else {
			comparison.PercentDelta = comparison.AbsoluteDelta / baseline * 100
			comparison.Classification = "stable"
			if math.Abs(comparison.PercentDelta) > signal.TolerancePct {
				if comparison.PercentDelta > 0 {
					comparison.Classification = "degraded"
				} else {
					comparison.Classification = "improved"
				}
			}
		}
		result = append(result, comparison)
	}
	return result, nil
}

func (s Source) averageRange(ctx context.Context, query string, start, end time.Time) (float64, error) {
	endpoint, err := url.Parse(strings.TrimRight(s.BaseURL, "/") + "/api/v1/query_range")
	if err != nil {
		return 0, err
	}
	step := s.Step
	if step <= 0 {
		step = time.Minute
	}
	params := endpoint.Query()
	params.Set("query", query)
	params.Set("start", strconv.FormatFloat(float64(start.UnixMilli())/1000, 'f', 3, 64))
	params.Set("end", strconv.FormatFloat(float64(end.UnixMilli())/1000, 'f', 3, 64))
	params.Set("step", strconv.FormatFloat(step.Seconds(), 'f', -1, 64))
	endpoint.RawQuery = params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return 0, err
	}
	if s.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+s.BearerToken)
	} else if s.BasicUser != "" {
		req.SetBasicAuth(s.BasicUser, s.BasicPass)
	}
	client := s.Client
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	response, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
		return 0, fmt.Errorf("Prometheus returned %s: %s", response.Status, strings.TrimSpace(string(body)))
	}
	var payload queryResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return 0, err
	}
	if payload.Status != "success" {
		return 0, fmt.Errorf("Prometheus query failed: %s", payload.Error)
	}
	var sum float64
	var count int
	for _, series := range payload.Data.Result {
		for _, sample := range series.Values {
			if len(sample) != 2 {
				continue
			}
			var rawValue string
			if err := json.Unmarshal(sample[1], &rawValue); err != nil {
				continue
			}
			value, err := strconv.ParseFloat(rawValue, 64)
			if err == nil && !math.IsNaN(value) && !math.IsInf(value, 0) {
				sum += value
				count++
			}
		}
	}
	if count == 0 {
		return 0, fmt.Errorf("no numeric samples")
	}
	return sum / float64(count), nil
}

type queryResponse struct {
	Status string `json:"status"`
	Error  string `json:"error"`
	Data   struct {
		Result []struct {
			Values [][]json.RawMessage `json:"values"`
		} `json:"result"`
	} `json:"data"`
}
