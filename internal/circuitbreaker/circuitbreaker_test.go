package circuitbreaker_test

import (
	"errors"
	"testing"
	"time"

	"edgecore/internal/circuitbreaker"
)

func TestCircuitBreaker_StateTransitions(t *testing.T) {
	cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
		Name:             "test-breaker",
		MaxRequests:      2,
		Interval:         10 * time.Second,
		Timeout:          100 * time.Millisecond,
		FailureThreshold: 2,
	})

	if cb.GetState() != circuitbreaker.StateClosed {
		t.Fatalf("expected initial state CLOSED, got %s", cb.GetState())
	}

	dummyErr := errors.New("upstream failed")

	// Trigger 1st failure
	_ = cb.Execute(func() error { return dummyErr })
	if cb.GetState() != circuitbreaker.StateClosed {
		t.Errorf("expected state CLOSED after 1 failure, got %s", cb.GetState())
	}

	// Trigger 2nd failure -> should trip to OPEN
	_ = cb.Execute(func() error { return dummyErr })
	if cb.GetState() != circuitbreaker.StateOpen {
		t.Fatalf("expected state OPEN after 2 failures, got %s", cb.GetState())
	}

	// Subsequent call should fail immediately with ErrCircuitBreakerOpen
	err := cb.Execute(func() error { return nil })
	if err != circuitbreaker.ErrCircuitBreakerOpen {
		t.Errorf("expected ErrCircuitBreakerOpen, got %v", err)
	}

	// Wait for timeout to transition to HALF_OPEN
	time.Sleep(150 * time.Millisecond)

	// Next call in HALF_OPEN should pass through
	successCalled := false
	err = cb.Execute(func() error {
		successCalled = true
		return nil
	})
	if err != nil || !successCalled {
		t.Errorf("expected call in HALF_OPEN to succeed, got err=%v", err)
	}

	// Second success in HALF_OPEN should transition back to CLOSED
	_ = cb.Execute(func() error { return nil })
	if cb.GetState() != circuitbreaker.StateClosed {
		t.Errorf("expected state CLOSED after recovery, got %s", cb.GetState())
	}
}
