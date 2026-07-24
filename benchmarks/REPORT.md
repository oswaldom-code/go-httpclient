# HTTP client comparison report

Generated: 2026-07-24 17:43 CEST

## Environment

| | |
|---|---|
| CPU | 12th Gen Intel(R) Core(TM) i7-1255U (12 threads) |
| OS/arch | linux/amd64 |
| Go | go1.24.1 |
| Samples per benchmark | 5 (min reported as typical cost) |

## Tool versions

- github.com/oswaldom-code/rhttp (local, via replace)
- github.com/go-resty/resty/v2 v2.17.2
- github.com/hashicorp/go-retryablehttp v0.7.8
- github.com/gojek/heimdall/v7 v7.0.3

## Methodology

All clients are configured equivalently: 5s timeout, 3 total attempts, exponential backoff 100ms-2s. Every client fully consumes and closes the response body.

- **Overhead**: a no-op transport returns 200 OK without touching the network, isolating client/middleware cost per request.
- **E2E**: a local httptest.Server returns ~1 KB of JSON over loopback, measuring total request cost including a real HTTP round trip.

Caveats: net/http does not retry (it is the floor, not a symmetric competitor); Heimdall runs without its Hystrix circuit breaker (retry only, for feature symmetry); Resty buffers the full response body by design; loopback amplifies relative overhead — against a real network (0.5-500 ms) these differences are negligible.

## Results: wrapper overhead (no network)

| Client | ns/op (min) | ns/op (mean) | B/op | allocs/op | vs best |
|---|---:|---:|---:|---:|---:|
| rhttp (Timeout+Retry+CircuitBreaker) | 1120 | 1176 | 1980 | 15 | 1.00x |
| rhttp (Timeout+Retry) | 1171 | 1231 | 1980 | 15 | 1.05x |
| go-retryablehttp | 1697 | 1750 | 1595 | 26 | 1.52x |
| net/http (Timeout only, no retry) | 1731 | 1772 | 1594 | 26 | 1.55x |
| Heimdall (retry) | 2283 | 2388 | 2220 | 32 | 2.04x |
| Resty (retry) | 5896 | 6307 | 4885 | 48 | 5.26x |

## Results: end-to-end (loopback, ~1 KB JSON)

| Client | ns/op (min) | ns/op (mean) | B/op | allocs/op | vs best |
|---|---:|---:|---:|---:|---:|
| go-retryablehttp | 57732 | 62533 | 6318 | 74 | 1.00x |
| net/http (Timeout only, no retry) | 58404 | 61336 | 6399 | 75 | 1.01x |
| rhttp (Timeout+Retry+CircuitBreaker) | 59532 | 63280 | 7596 | 79 | 1.03x |
| Heimdall (retry) | 60432 | 65872 | 6953 | 80 | 1.05x |
| rhttp (Timeout+Retry) | 64372 | 67119 | 7567 | 79 | 1.12x |
| Resty (retry) | 72257 | 79984 | 10717 | 96 | 1.25x |

## Reproduce

```bash
cd benchmarks
go run ./report -count 5
```
