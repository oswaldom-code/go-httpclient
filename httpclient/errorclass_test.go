package httpclient_test

import (
	"context"
	"errors"
	"net"
	"net/url"
	"testing"

	"github.com/oswaldom-code/go-httpclient/httpclient"
)

func TestClassify_DeadlineExceeded(t *testing.T) {
	classified := httpclient.Classify(context.DeadlineExceeded)

	if classified.Kind != httpclient.ErrKindTimeout {
		t.Errorf("expected ErrKindTimeout, got %v", classified.Kind)
	}
	if !errors.Is(classified, context.DeadlineExceeded) {
		t.Error("expected Unwrap to return original error")
	}
}

func TestClassify_Canceled(t *testing.T) {
	classified := httpclient.Classify(context.Canceled)

	if classified.Kind != httpclient.ErrKindCanceled {
		t.Errorf("expected ErrKindCanceled, got %v", classified.Kind)
	}
}

func TestClassify_DNSError(t *testing.T) {
	dnsErr := &net.DNSError{
		Err:  "no such host",
		Name: "invalid.example.com",
	}
	classified := httpclient.Classify(dnsErr)

	if classified.Kind != httpclient.ErrKindDNS {
		t.Errorf("expected ErrKindDNS, got %v", classified.Kind)
	}
}

func TestClassify_ConnectionRefused(t *testing.T) {
	err := errors.New("dial tcp 127.0.0.1:8080: connection refused")
	classified := httpclient.Classify(err)

	if classified.Kind != httpclient.ErrKindConnection {
		t.Errorf("expected ErrKindConnection, got %v", classified.Kind)
	}
}

func TestClassify_ConnectionReset(t *testing.T) {
	err := errors.New("read tcp: connection reset by peer")
	classified := httpclient.Classify(err)

	if classified.Kind != httpclient.ErrKindConnection {
		t.Errorf("expected ErrKindConnection, got %v", classified.Kind)
	}
}

func TestClassify_TLSError(t *testing.T) {
	err := errors.New("tls: certificate signed by unknown authority")
	classified := httpclient.Classify(err)

	if classified.Kind != httpclient.ErrKindTLS {
		t.Errorf("expected ErrKindTLS, got %v", classified.Kind)
	}
}

func TestClassify_X509Error(t *testing.T) {
	err := errors.New("x509: certificate has expired")
	classified := httpclient.Classify(err)

	if classified.Kind != httpclient.ErrKindTLS {
		t.Errorf("expected ErrKindTLS, got %v", classified.Kind)
	}
}

func TestClassify_WrappedURLError(t *testing.T) {
	urlErr := &url.Error{
		Op:  "Get",
		URL: "http://example.com",
		Err: context.DeadlineExceeded,
	}
	classified := httpclient.Classify(urlErr)

	if classified.Kind != httpclient.ErrKindTimeout {
		t.Errorf("expected ErrKindTimeout for wrapped deadline, got %v", classified.Kind)
	}
}

func TestClassify_URLErrorTimeout(t *testing.T) {
	urlErr := &url.Error{
		Op:  "Get",
		URL: "http://example.com",
		Err: &timeoutError{},
	}
	classified := httpclient.Classify(urlErr)

	if classified.Kind != httpclient.ErrKindTimeout {
		t.Errorf("expected ErrKindTimeout, got %v", classified.Kind)
	}
}

// timeoutError implements net.Error with Timeout() = true
type timeoutError struct{}

func (e *timeoutError) Error() string   { return "timeout" }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return true }

func TestClassify_NilError(t *testing.T) {
	classified := httpclient.Classify(nil)

	if classified != nil {
		t.Error("expected nil for nil error")
	}
}

func TestClassify_UnknownError(t *testing.T) {
	err := errors.New("something completely unexpected")
	classified := httpclient.Classify(err)

	if classified.Kind != httpclient.ErrKindUnknown {
		t.Errorf("expected ErrKindUnknown, got %v", classified.Kind)
	}
}

func TestClassifiedError_Error(t *testing.T) {
	err := errors.New("connection refused")
	classified := httpclient.Classify(err)

	expected := "connection: connection refused"
	if classified.Error() != expected {
		t.Errorf("expected %q, got %q", expected, classified.Error())
	}
}

func TestClassifiedError_Unwrap(t *testing.T) {
	originalErr := errors.New("original error")
	classified := httpclient.Classify(originalErr)

	if !errors.Is(classified, originalErr) {
		t.Error("errors.Is should match original error")
	}
}

func TestErrorKind_String(t *testing.T) {
	tests := []struct {
		kind     httpclient.ErrorKind
		expected string
	}{
		{httpclient.ErrKindTimeout, "timeout"},
		{httpclient.ErrKindCanceled, "canceled"},
		{httpclient.ErrKindConnection, "connection"},
		{httpclient.ErrKindDNS, "dns"},
		{httpclient.ErrKindTLS, "tls"},
		{httpclient.ErrKindTemporary, "temporary"},
		{httpclient.ErrKindUnknown, "unknown"},
	}

	for _, tt := range tests {
		if tt.kind.String() != tt.expected {
			t.Errorf("expected %q, got %q", tt.expected, tt.kind.String())
		}
	}
}

func TestErrorKind_IsRetryable(t *testing.T) {
	retryable := []httpclient.ErrorKind{
		httpclient.ErrKindTimeout,
		httpclient.ErrKindConnection,
		httpclient.ErrKindDNS,
		httpclient.ErrKindTemporary,
	}
	for _, k := range retryable {
		if !k.IsRetryable() {
			t.Errorf("expected %v to be retryable", k)
		}
	}

	notRetryable := []httpclient.ErrorKind{
		httpclient.ErrKindCanceled,
		httpclient.ErrKindTLS,
		httpclient.ErrKindUnknown,
	}
	for _, k := range notRetryable {
		if k.IsRetryable() {
			t.Errorf("expected %v to not be retryable", k)
		}
	}
}

func TestIsTimeout(t *testing.T) {
	if !httpclient.IsTimeout(context.DeadlineExceeded) {
		t.Error("expected IsTimeout to be true for DeadlineExceeded")
	}
	if httpclient.IsTimeout(context.Canceled) {
		t.Error("expected IsTimeout to be false for Canceled")
	}
	if httpclient.IsTimeout(nil) {
		t.Error("expected IsTimeout to be false for nil")
	}
}

func TestIsCanceled(t *testing.T) {
	if !httpclient.IsCanceled(context.Canceled) {
		t.Error("expected IsCanceled to be true for Canceled")
	}
	if httpclient.IsCanceled(context.DeadlineExceeded) {
		t.Error("expected IsCanceled to be false for DeadlineExceeded")
	}
	if httpclient.IsCanceled(nil) {
		t.Error("expected IsCanceled to be false for nil")
	}
}

func TestIsConnection(t *testing.T) {
	err := errors.New("connection refused")
	if !httpclient.IsConnection(err) {
		t.Error("expected IsConnection to be true for connection refused")
	}
	if httpclient.IsConnection(context.Canceled) {
		t.Error("expected IsConnection to be false for Canceled")
	}
}

func TestIsDNS(t *testing.T) {
	dnsErr := &net.DNSError{Err: "no such host", Name: "invalid.example.com"}
	if !httpclient.IsDNS(dnsErr) {
		t.Error("expected IsDNS to be true for DNSError")
	}
	if httpclient.IsDNS(context.Canceled) {
		t.Error("expected IsDNS to be false for Canceled")
	}
}

func TestIsTLS(t *testing.T) {
	err := errors.New("tls: handshake failure")
	if !httpclient.IsTLS(err) {
		t.Error("expected IsTLS to be true for TLS error")
	}
	if httpclient.IsTLS(context.Canceled) {
		t.Error("expected IsTLS to be false for Canceled")
	}
}

func TestIsRetryable(t *testing.T) {
	// Retryable
	if !httpclient.IsRetryable(context.DeadlineExceeded) {
		t.Error("expected timeout to be retryable")
	}
	if !httpclient.IsRetryable(errors.New("connection refused")) {
		t.Error("expected connection error to be retryable")
	}

	// Not retryable
	if httpclient.IsRetryable(context.Canceled) {
		t.Error("expected canceled to not be retryable")
	}
	if httpclient.IsRetryable(errors.New("tls: certificate error")) {
		t.Error("expected TLS error to not be retryable")
	}
	if httpclient.IsRetryable(nil) {
		t.Error("expected nil to not be retryable")
	}
}
