package http

import "fmt"

// Error wraps transport or validation failures for callers.
type Error struct {
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("http error: %s", e.Message)
}
