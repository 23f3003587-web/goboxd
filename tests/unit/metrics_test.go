package unit_test

import (
	"testing"

	"github.com/thesouldev/goboxd/internal/metrics"
)

func TestMetricsCountersTrackRequestsAndFailures(t *testing.T) {
	beforeRequests := metrics.Requests()
	metrics.IncRequests()
	if got := metrics.Requests(); got != beforeRequests+1 {
		t.Fatalf("expected request count to increase by 1, got %d", got)
	}

	beforeInFlight := metrics.InFlight()
	metrics.IncInFlight()
	if got := metrics.InFlight(); got != beforeInFlight+1 {
		t.Fatalf("expected in-flight count to increase by 1, got %d", got)
	}

	metrics.DecInFlight()
	if got := metrics.InFlight(); got != beforeInFlight {
		t.Fatalf("expected in-flight count to return to baseline %d, got %d", beforeInFlight, got)
	}

	beforeFailed := metrics.FailedInternal()
	metrics.IncFailedInternal()
	if got := metrics.FailedInternal(); got != beforeFailed+1 {
		t.Fatalf("expected failed count to increase by 1, got %d", got)
	}

	metrics.SetLastInternalErrorAt("2026-06-10T12:00:00Z")
	if got := metrics.LastInternalErrorAt(); got != "2026-06-10T12:00:00Z" {
		t.Fatalf("expected recorded timestamp, got %q", got)
	}
}
