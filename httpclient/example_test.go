package httpclient_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/oswaldom-code/go-httpclient/httpclient"
)

func ExampleNew() {
	// Create a basic client with default settings
	client := httpclient.New()

	req, _ := http.NewRequest("GET", "https://api.example.com/users", http.NoBody)
	resp, err := client.Do(context.Background(), req)
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)
}

func ExampleNew_withMiddleware() {
	// Create a client with timeout, retry, and circuit breaker
	client := httpclient.New(
		httpclient.WithMiddleware(
			httpclient.Timeout(5*time.Second),
			httpclient.Retry(httpclient.RetryConfig{
				MaxAttempts: 3,
				Backoff:     httpclient.ExponentialBackoff(100*time.Millisecond, 5*time.Second),
			}),
			httpclient.CircuitBreaker(httpclient.CircuitBreakerConfig{
				FailureThreshold: 5,
				ResetTimeout:     30 * time.Second,
			}),
		),
	)

	req, _ := http.NewRequest("GET", "https://api.example.com/users", http.NoBody)
	resp, err := client.Do(context.Background(), req)
	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)
}

func ExampleR() {
	client := httpclient.New()

	// Use the fluent API to build and execute requests
	resp, err := httpclient.R(client).
		SetHeader("Authorization", "Bearer token").
		SetQueryParam("page", "1").
		Get("https://api.example.com/users")

	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)
}

func ExampleRequestBuilder_SetBodyJSON() {
	client := httpclient.New()

	type User struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	user := User{Name: "John", Email: "john@example.com"}

	resp, err := httpclient.R(client).
		SetBodyJSON(user).
		Post("https://api.example.com/users")

	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Status:", resp.StatusCode)
}

func ExampleRequestBuilder_SetPathParam() {
	client := httpclient.New()

	// Path parameters are replaced in the URL template
	resp, err := httpclient.R(client).
		SetPathParam("id", "123").
		Get("https://api.example.com/users/{id}")

	if err != nil {
		fmt.Println("request failed:", err)
		return
	}
	defer resp.Body.Close()

	// Request was made to: https://api.example.com/users/123
	fmt.Println("Status:", resp.StatusCode)
}

func ExampleClassify() {
	client := httpclient.New(
		httpclient.WithMiddleware(
			httpclient.Timeout(100 * time.Millisecond),
		),
	)

	req, _ := http.NewRequest("GET", "https://slow-api.example.com", http.NoBody)
	_, err := client.Do(context.Background(), req)

	if err != nil {
		classified := httpclient.Classify(err)
		fmt.Printf("Error kind: %s, Retryable: %v\n",
			classified.Kind, classified.Kind.IsRetryable())
	}
}

func ExampleExponentialBackoff() {
	backoff := httpclient.ExponentialBackoff(100*time.Millisecond, 10*time.Second)

	// Backoff durations increase exponentially with jitter
	fmt.Println("Attempt 0:", backoff(0)) // ~100ms
	fmt.Println("Attempt 1:", backoff(1)) // ~200ms
	fmt.Println("Attempt 2:", backoff(2)) // ~400ms
}

func ExampleNewTokenBucket() {
	// Allow 10 requests per second with burst of 5
	limiter := httpclient.NewTokenBucket(10, 5)

	// Use with rate limit middleware
	client := httpclient.New(
		httpclient.WithMiddleware(
			httpclient.RateLimit(httpclient.RateLimitConfig{
				Limiter:     limiter,
				WaitOnLimit: true,
			}),
		),
	)

	_ = client // Use client for requests
}

func ExampleCircuitBreaker() {
	// Circuit breaker opens after 5 failures
	// and stays open for 30 seconds before trying again
	client := httpclient.New(
		httpclient.WithMiddleware(
			httpclient.CircuitBreaker(httpclient.CircuitBreakerConfig{
				FailureThreshold: 5,
				ResetTimeout:     30 * time.Second,
			}),
		),
	)

	_ = client // Use client for requests
}

func ExampleLogging() {
	// Custom logger that prints request/response details
	logger := httpclient.LoggerFunc(func(entry httpclient.LogEntry) {
		fmt.Printf("%s %s -> %d (%s)\n",
			entry.Method, entry.URL, entry.StatusCode, entry.Duration)
	})

	client := httpclient.New(
		httpclient.WithMiddleware(
			httpclient.Logging(httpclient.LoggingConfig{
				Logger: logger,
			}),
		),
	)

	_ = client // Use client for requests
}

func ExampleGetBuffer() {
	// Get a buffer from the pool
	buf := httpclient.GetBuffer()

	// Use the buffer
	buf.WriteString("Hello, World!")

	// Return to pool when done
	httpclient.PutBuffer(buf)
}
