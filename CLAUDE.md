# CLAUDE.md - Project Instructions

## Description
Production-grade HTTP client for Go with built-in resiliency patterns. Zero external dependencies.

## Project Structure

```
rhttp/                    # Package rhttp lives at the module root
├── client.go             # Client struct and New() constructor
├── middleware.go         # Middleware and RoundTripperFunc types, chain()
├── transport.go          # Optimized DefaultTransport()
├── options.go            # Functional options pattern
├── errors.go             # Sentinel errors
├── errorclass.go         # Error classification
├── timeout.go            # Timeout middleware
├── retry.go              # Retry middleware
├── backoff.go            # Backoff strategies (7 variants)
├── circuitbreaker.go     # Thread-safe circuit breaker
├── ratelimit.go          # Token bucket rate limiter
├── logging.go            # Logging middleware
├── metrics.go            # Metrics middleware
├── request.go            # Fluent API (RequestBuilder)
├── pool.go               # Object pooling with sync.Pool
├── examples/             # Runnable examples
├── .github/workflows/    # CI with GitHub Actions
├── Makefile              # Development commands
└── .golangci.yml         # Linter configuration
```

## Development Commands

```bash
make test          # Run tests
make test-race     # Run tests with race detector
make test-coverage # Generate coverage report
make coverage-summary # Show coverage summary
make bench         # Run benchmarks
make lint          # Run golangci-lint
make check         # Run all checks
make fmt           # Format code
```

## Code Conventions

### Middleware
- Implement as `func(http.RoundTripper) http.RoundTripper`
- Use struct implementing `RoundTrip(req *http.Request) (*http.Response, error)`
- If config is nil or invalid, return next unchanged

```go
func MyMiddleware(cfg Config) Middleware {
    if cfg.Invalid() {
        return func(next http.RoundTripper) http.RoundTripper {
            return next
        }
    }
    return func(next http.RoundTripper) http.RoundTripper {
        return myRoundTripper{next: next, cfg: cfg}
    }
}
```

### Tests
- Use standard `testing` package (project does NOT use ginkgo/gomega by design - zero deps)
- Name files `*_test.go`
- Use `rhttp.RoundTripperFunc` for transport mocks
- Respect context in mocks with `select { case <-req.Context().Done(): ... }`

### Errors
- Sentinel errors in `errors.go`: `var ErrXxx = errors.New("rhttp: description")`
- Error classification in `errorclass.go`

### Backoff
- Functions returning `BackoffFunc = func(attempt int) time.Duration`
- Zero allocations (verify with benchmarks)
- Respect max duration

## Design Patterns Used

1. **Middleware Chain** - Chained RoundTrippers
2. **Functional Options** - Configuration with `WithXxx()`
3. **Circuit Breaker** - State machine (Closed/Open/Half-Open)
4. **Token Bucket** - Rate limiting
5. **Object Pool** - sync.Pool for buffers
6. **Builder Pattern** - Fluent API in RequestBuilder

## Recommended Middleware Order

```go
Logging → Metrics → Timeout → RateLimit → Retry → CircuitBreaker
```

Retry sits outside CircuitBreaker so every attempt consults the circuit: a
tripped breaker short-circuits the remaining attempts.

## Pre-Commit Checklist

1. `make fmt` - Code formatted
2. `make lint` - No linter errors
3. `make test-race` - Tests pass with race detector
4. `go mod tidy` - go.mod is clean

## Performance

- Client must be faster than standard `net/http`
- Rate limiter: ~50ns per operation, zero allocs
- Backoff strategies: <10ns, zero allocs
- Run `make bench` to check for regressions

## Pending Roadmap

See "Roadmap" section in README.md for pending features:
- Circuit breaker per endpoint
- OpenTelemetry integration
- OAuth2 support
- Load balancing
- Response caching
