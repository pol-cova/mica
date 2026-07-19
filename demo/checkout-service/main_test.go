package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func resetDemoMetrics() {
	regression.Store(false)
	calls.Store(0)
	metrics.requestsOK.Store(0)
	metrics.requests5xx.Store(0)
	metrics.queries.Store(0)
	for index := range metrics.buckets {
		metrics.buckets[index].Store(0)
	}
}

func TestCheckoutUsesBatchLookupUntilRegressionIsEnabled(t *testing.T) {
	resetDemoMetrics()
	t.Cleanup(resetDemoMetrics)

	healthy := httptest.NewRecorder()
	checkout(healthy, httptest.NewRequest(http.MethodGet, "/checkout?items=8", nil))
	if healthy.Code != http.StatusOK {
		t.Fatalf("healthy status = %d", healthy.Code)
	}
	if got := metrics.queries.Load(); got != 2 {
		t.Fatalf("healthy query count = %d, want batch lookup plus order query", got)
	}

	regression.Store(true)
	degraded := httptest.NewRecorder()
	checkout(degraded, httptest.NewRequest(http.MethodGet, "/checkout?items=8", nil))
	if degraded.Code != http.StatusOK {
		t.Fatalf("regression status = %d", degraded.Code)
	}
	if got := metrics.queries.Load(); got != 11 {
		t.Fatalf("query count = %d, want 2 healthy + 9 N+1 queries", got)
	}
}
