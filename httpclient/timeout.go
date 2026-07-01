package httpclient

import (
	"context"
	"net/http"
	"time"
)

// Timeout returns a middleware that applies a timeout to requests.
// If the request's context already has a shorter deadline, it is respected.
func Timeout(d time.Duration) Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return timeoutRoundTripper{next: next, timeout: d}
	}
}

type timeoutRoundTripper struct {
	next    http.RoundTripper
	timeout time.Duration
}

func (t timeoutRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// Only apply timeout if ctx doesn't have a shorter deadline
	if deadline, ok := ctx.Deadline(); ok {
		if time.Until(deadline) <= t.timeout {
			return t.next.RoundTrip(req)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	req = req.Clone(ctx)
	return t.next.RoundTrip(req)
}
