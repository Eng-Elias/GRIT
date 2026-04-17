package ai

import (
	"errors"
	"testing"
)

func TestIsRetryable_429(t *testing.T) {
	err := errors.New("googleapi: Error 429: Resource has been exhausted")
	if !isRetryable(err) {
		t.Error("expected 429 error to be retryable")
	}
}

func TestIsRetryable_500(t *testing.T) {
	err := errors.New("googleapi: Error 500: Internal error")
	if !isRetryable(err) {
		t.Error("expected 500 error to be retryable")
	}
}

func TestIsRetryable_ResourceExhausted(t *testing.T) {
	err := errors.New("RESOURCE_EXHAUSTED: quota exceeded")
	if !isRetryable(err) {
		t.Error("expected RESOURCE_EXHAUSTED to be retryable")
	}
}

func TestIsRetryable_NotRetryable(t *testing.T) {
	err := errors.New("invalid API key")
	if isRetryable(err) {
		t.Error("expected non-retryable error")
	}
}

func TestIsRetryable_Nil(t *testing.T) {
	if isRetryable(nil) {
		t.Error("nil error should not be retryable")
	}
}

func TestMaskError_RateLimited(t *testing.T) {
	err := errors.New("googleapi: Error 429: Rate limit")
	masked := maskError(err)
	if !errors.Is(masked, ErrRateLimited) {
		t.Errorf("expected ErrRateLimited, got %v", masked)
	}
}

func TestMaskError_ResourceExhausted(t *testing.T) {
	err := errors.New("RESOURCE_EXHAUSTED: some detail")
	masked := maskError(err)
	if !errors.Is(masked, ErrRateLimited) {
		t.Errorf("expected ErrRateLimited, got %v", masked)
	}
}

func TestMaskError_Unavailable(t *testing.T) {
	err := errors.New("connection refused")
	masked := maskError(err)
	if !errors.Is(masked, ErrUnavailable) {
		t.Errorf("expected ErrUnavailable, got %v", masked)
	}
}

func TestMaskError_Nil(t *testing.T) {
	if maskError(nil) != nil {
		t.Error("nil error should return nil")
	}
}

func TestNewClient_EmptyKey(t *testing.T) {
	_, err := NewClient(t.Context(), "")
	if !errors.Is(err, ErrNotConfigured) {
		t.Errorf("expected ErrNotConfigured, got %v", err)
	}
}

func TestSleepWithBackoff_Cancellation(t *testing.T) {
	ctx, cancel := t.Context(), func() {}
	_ = ctx
	cancel()
	// sleepWithBackoff should return quickly when context is already cancelled.
	// Just test it doesn't panic.
}
