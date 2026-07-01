package httpclient

import "errors"

var (
	// ErrInvalidRequest is returned when a nil request is passed to Do.
	ErrInvalidRequest = errors.New("httpclient: invalid request")

	// ErrCircuitOpen is returned when the circuit breaker is open.
	ErrCircuitOpen = errors.New("httpclient: circuit breaker is open")

	// ErrRateLimited is returned when the rate limit is exceeded and WaitOnLimit is false.
	ErrRateLimited = errors.New("httpclient: rate limit exceeded")
)
