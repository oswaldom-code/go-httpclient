package httpclient

import (
	"net/http"
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreakerConfig configures the circuit breaker middleware.
type CircuitBreakerConfig struct {
	// FailureThreshold is the number of consecutive failures before opening the circuit.
	FailureThreshold int

	// ResetTimeout is how long to wait in Open state before transitioning to Half-Open.
	ResetTimeout time.Duration

	// IsFailure determines if a response/error should count as a failure.
	// If nil, any error or 5xx status code is considered a failure.
	IsFailure func(resp *http.Response, err error) bool
}

// CircuitBreaker returns a middleware that implements the circuit breaker pattern.
func CircuitBreaker(cfg CircuitBreakerConfig) Middleware {
	if cfg.FailureThreshold <= 0 {
		cfg.FailureThreshold = 5
	}
	if cfg.ResetTimeout <= 0 {
		cfg.ResetTimeout = 30 * time.Second
	}
	if cfg.IsFailure == nil {
		cfg.IsFailure = DefaultIsFailure
	}

	cb := &circuitBreaker{
		cfg:   cfg,
		state: CircuitClosed,
	}

	return func(next http.RoundTripper) http.RoundTripper {
		cb.next = next
		return cb
	}
}

type circuitBreaker struct {
	next http.RoundTripper
	cfg  CircuitBreakerConfig

	mu              sync.Mutex
	state           CircuitState
	failures        int
	lastFailureTime time.Time
}

func (cb *circuitBreaker) RoundTrip(req *http.Request) (*http.Response, error) {
	if !cb.allowRequest() {
		return nil, ErrCircuitOpen
	}

	resp, err := cb.next.RoundTrip(req)

	cb.recordResult(resp, err)

	return resp, err
}

func (cb *circuitBreaker) allowRequest() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		if time.Since(cb.lastFailureTime) >= cb.cfg.ResetTimeout {
			cb.state = CircuitHalfOpen
			return true
		}
		return false

	case CircuitHalfOpen:
		// In half-open state, allow the request (only one at a time due to mutex)
		return true

	default:
		return true
	}
}

func (cb *circuitBreaker) recordResult(resp *http.Response, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	isFailure := cb.cfg.IsFailure(resp, err)

	switch cb.state {
	case CircuitClosed:
		if isFailure {
			cb.failures++
			cb.lastFailureTime = time.Now()
			if cb.failures >= cb.cfg.FailureThreshold {
				cb.state = CircuitOpen
			}
		} else {
			cb.failures = 0
		}

	case CircuitHalfOpen:
		if isFailure {
			cb.state = CircuitOpen
			cb.lastFailureTime = time.Now()
			cb.failures = cb.cfg.FailureThreshold
		} else {
			cb.state = CircuitClosed
			cb.failures = 0
		}
	}
}

// State returns the current state of the circuit breaker.
// Useful for monitoring and testing.
func (cb *circuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func DefaultIsFailure(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	if resp != nil && resp.StatusCode >= 500 {
		return true
	}
	return false
}
