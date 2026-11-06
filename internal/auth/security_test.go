package auth

import (
	"io"
	"log/slog"
	"testing"
	"time"
)

func TestProgressiveDelayCannotOverflow(t *testing.T) {
	for _, failures := range []int{1, 2, 5, 6, 34, 64, 65, 100} {
		tracker := &FailedLoginTracker{attempts: make(map[string]*loginAttempt)}
		for i := 0; i < failures; i++ {
			tracker.RecordFailure("test-client")
		}
		expected := 30 * time.Second
		if failures < 6 {
			expected = time.Duration(1<<(failures-1)) * time.Second
		}
		if got := tracker.GetDelay("test-client"); got < expected-time.Second || got > expected {
			t.Errorf("after %d failures: expected near %v, got %v", failures, expected, got)
		}
		tracker.attempts["test-client"].lastFail = time.Now().Add(-31 * time.Second)
		if got := tracker.GetDelay("test-client"); got != 0 {
			t.Errorf("elapsed delay: %v", got)
		}
		tracker.RecordSuccess("test-client")
		if got := tracker.GetDelay("test-client"); got != 0 {
			t.Errorf("success did not reset: %v", got)
		}
	}
}

func TestConcurrentFailuresCannotSkipProgressiveDelay(t *testing.T) {
	svc := NewService(nil, "correct-password", "fixture-secret", time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)), false)
	const requests = 4 // Small, bounded probe; never stress or exhaust the machine.
	start := make(chan struct{})
	results := make(chan time.Duration, requests)
	for i := 0; i < requests; i++ {
		go func() {
			<-start
			_, delay := svc.VerifyPasswordWithDelay("wrong-password", "same-client")
			results <- delay
		}()
	}
	close(start)
	withoutDelay := 0
	for i := 0; i < requests; i++ {
		if <-results == 0 {
			withoutDelay++
		}
	}
	if withoutDelay != 1 {
		t.Errorf("concurrent attempts bypassed delay: %d expensive checks admitted, want 1", withoutDelay)
	}
}

func TestBusyPasswordVerificationReturnsRetryWithoutRecordingFailure(t *testing.T) {
	svc := NewService(nil, "correct-password", "fixture-secret", time.Hour, slog.New(slog.NewTextHandler(io.Discard, nil)), false)
	svc.passwordMu.Lock()
	valid, delay := svc.VerifyPasswordWithDelay("correct-password", "other-client")
	svc.passwordMu.Unlock()
	if valid || delay <= 0 {
		t.Fatalf("busy verification: valid=%v delay=%v", valid, delay)
	}
	if delay := svc.failedAttempts.GetDelay("other-client"); delay != 0 {
		t.Fatalf("busy attempt recorded a password failure: %v", delay)
	}
	if valid, delay := svc.VerifyPasswordWithDelay("correct-password", "other-client"); !valid || delay != 0 {
		t.Fatalf("correct password after busy: valid=%v delay=%v", valid, delay)
	}
	if !svc.VerifyPassword("correct-password") || svc.VerifyPassword("wrong-password") {
		t.Fatal("direct password verification changed")
	}
}
