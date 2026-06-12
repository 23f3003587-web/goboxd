package handlers

import (
	"context"
	"sync"

	"github.com/thesouldev/goboxd/internal/config"
	"github.com/thesouldev/goboxd/internal/metrics"
)

var (
	concurrencyLimiter  chan struct{}
	concurrencyInitOnce sync.Once
)

func InitializeConcurrency() {
	concurrencyInitOnce.Do(func() {
		limit := config.Global.MaxConcurrentJobs
		if limit <= 0 {
			limit = 1
		}
		concurrencyLimiter = make(chan struct{}, limit)
	})
}

func AcquireJob(ctx context.Context) error {
	InitializeConcurrency()

	select {
	case concurrencyLimiter <- struct{}{}:
		metrics.IncInFlight()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func ReleaseJob() {
	if concurrencyLimiter == nil {
		return
	}
	select {
	case <-concurrencyLimiter:
		metrics.DecInFlight()
	default:
	}
}
