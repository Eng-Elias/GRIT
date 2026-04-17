package ai

import (
	"testing"
)

func TestRateLimiter_AllowWithinLimit(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	for i := range chatRateLimit {
		if !rl.Allow("192.168.1.1") {
			t.Fatalf("request %d should be allowed within limit", i+1)
		}
	}
}

func TestRateLimiter_DenyOverLimit(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	// Exhaust the burst.
	for range chatBurstLimit {
		rl.Allow("10.0.0.1")
	}

	// Next request should be denied.
	if rl.Allow("10.0.0.1") {
		t.Error("request should be denied after exceeding limit")
	}
}

func TestRateLimiter_PerIPIsolation(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	// Exhaust IP-A.
	for range chatBurstLimit {
		rl.Allow("ip-a")
	}

	// IP-B should still be allowed.
	if !rl.Allow("ip-b") {
		t.Error("different IP should have independent limit")
	}
}

func TestRateLimiter_StopDoesNotPanic(t *testing.T) {
	rl := NewRateLimiter()
	rl.Stop()
	// Calling Allow after Stop should not panic.
	_ = rl.Allow("127.0.0.1")
}
