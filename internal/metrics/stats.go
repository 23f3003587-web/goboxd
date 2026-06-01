package metrics

import (
	"sync"
)

var (
	mu                sync.Mutex
	requests          int
	inFlight          int
	failedInternal    int
	lastInternalError string
)

// IncRequests increments the global request counter.
func IncRequests() {
	mu.Lock()
	requests++
	mu.Unlock()
}

// Requests returns the current number of requests recorded.
func Requests() int {
	mu.Lock()
	defer mu.Unlock()
	return requests
}

// IncInFlight increments the number of active jobs.
func IncInFlight() {
	mu.Lock()
	inFlight++
	mu.Unlock()
}

// DecInFlight decrements the number of active jobs.
func DecInFlight() {
	mu.Lock()
	if inFlight > 0 {
		inFlight--
	}
	mu.Unlock()
}

// InFlight returns the current number of active jobs.
func InFlight() int {
	mu.Lock()
	defer mu.Unlock()
	return inFlight
}

// IncFailedInternal increments the count of internal failures.
func IncFailedInternal() {
	mu.Lock()
	failedInternal++
	mu.Unlock()
}

// FailedInternal returns the count of internal failures.
func FailedInternal() int {
	mu.Lock()
	defer mu.Unlock()
	return failedInternal
}

// SetLastInternalErrorAt records the timestamp of the most recent internal failure.
func SetLastInternalErrorAt(timestamp string) {
	mu.Lock()
	lastInternalError = timestamp
	mu.Unlock()
}

// LastInternalErrorAt returns the timestamp of the last internal failure.
func LastInternalErrorAt() string {
	mu.Lock()
	defer mu.Unlock()
	return lastInternalError
}
