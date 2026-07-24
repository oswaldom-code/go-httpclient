# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

First public release, tagged as `v0.1.0`.

### Added

- Middleware-based HTTP client (`New`, `WithMiddleware`, `WithTransport`) built on `http.RoundTripper`.
- Resiliency middleware: `Timeout`, `Retry` with pluggable backoff, `CircuitBreaker`, and `RateLimit` (token bucket).
- `SharedCircuitBreaker` (`NewCircuitBreaker`) for circuit state shared across multiple clients.
- Observability middleware: `Logging` and `Metrics`. `MetricsConfig.PathNormalizer` bounds metrics label cardinality (raw path is omitted by default).
- Backoff strategies: constant, linear, exponential, Fibonacci, and their jitter variants.
- Fluent request builder (`R`) with JSON and reader bodies, path parameters, and query parameters. Reader bodies up to 10 MB are buffered so retries can rewind them.
- Error classification: `Classify`, `IsRetryable`, `IsTimeout`, `IsConnection`, and related helpers.
- Zero external dependencies; Go standard library only.

[Unreleased]: https://github.com/oswaldom-code/rhttp/commits/develop
