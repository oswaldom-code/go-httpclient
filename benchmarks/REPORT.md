# HTTP client comparison report

Generated: 2026-07-25 11:58 CEST

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
| rhttp (Timeout+Retry) | 1020 | 1100 | 1980 | 15 | 1.00x |
| rhttp (Timeout+Retry+CircuitBreaker) | 1102 | 1132 | 1980 | 15 | 1.08x |
| net/http (Timeout only, no retry) | 1713 | 1747 | 1594 | 26 | 1.68x |
| go-retryablehttp | 1728 | 1754 | 1594 | 26 | 1.69x |
| Heimdall (retry) | 2410 | 2493 | 2220 | 32 | 2.36x |
| Resty (retry) | 5834 | 6000 | 4885 | 48 | 5.72x |

## Results: end-to-end (loopback, ~1 KB JSON)

| Client | ns/op (min) | ns/op (mean) | B/op | allocs/op | vs best |
|---|---:|---:|---:|---:|---:|
| net/http (Timeout only, no retry) | 57001 | 64212 | 6502 | 75 | 1.00x |
| go-retryablehttp | 58000 | 60629 | 6314 | 74 | 1.02x |
| Heimdall (retry) | 58871 | 60608 | 6955 | 80 | 1.03x |
| rhttp (Timeout+Retry+CircuitBreaker) | 61003 | 66404 | 7652 | 79 | 1.07x |
| rhttp (Timeout+Retry) | 61185 | 64498 | 7576 | 79 | 1.07x |
| Resty (retry) | 67672 | 74658 | 10721 | 96 | 1.19x |

## Reproduce

```bash
cd benchmarks
go run ./report -count 5
```
