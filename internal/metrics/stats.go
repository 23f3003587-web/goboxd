package metrics

import "sync"

var (
	mu       sync.Mutex
	requests int
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
