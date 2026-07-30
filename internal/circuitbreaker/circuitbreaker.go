package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// CircuitBreakerState represents the current state of a circuit breaker
type State int

const (
	StateClosed State = iota
	StateHalfOpen
	StateOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateHalfOpen:
		return "HALF_OPEN"
	case StateOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

var ErrCircuitBreakerOpen = errors.New("circuit breaker is OPEN - upstream request rejected")

// Config settings for CircuitBreaker
type Config struct {
	Name             string
	MaxRequests      uint32        // Max requests allowed in Half-Open state
	Interval         time.Duration // Time window to reset failure counters
	Timeout          time.Duration // Time in Open state before transitioning to Half-Open
	FailureThreshold uint32        // Consecutive failures threshold to trip to Open
}

// CircuitBreaker state machine
type CircuitBreaker struct {
	name             string
	maxRequests      uint32
	interval         time.Duration
	timeout          time.Duration
	failureThreshold uint32

	mu          sync.Mutex
	state       State
	generation  uint64
	counts      Counts
	expiry      time.Time
}

// Counts records request statistics within current generation
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

func (c *Counts) onRequest() {
	c.Requests++
}

func (c *Counts) onSuccess() {
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
}

func (c *Counts) onFailure() {
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
}

func (c *Counts) clear() {
	c.Requests = 0
	c.TotalSuccesses = 0
	c.TotalFailures = 0
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
}

// NewCircuitBreaker initializes a circuit breaker instance
func NewCircuitBreaker(cfg Config) *CircuitBreaker {
	if cfg.MaxRequests == 0 {
		cfg.MaxRequests = 1
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}
	if cfg.FailureThreshold == 0 {
		cfg.FailureThreshold = 5
	}

	cb := &CircuitBreaker{
		name:             cfg.Name,
		maxRequests:      cfg.MaxRequests,
		interval:         cfg.Interval,
		timeout:          cfg.Timeout,
		failureThreshold: cfg.FailureThreshold,
	}
	cb.toNewGeneration(time.Now())

	return cb
}

// Execute runs the given function guarded by circuit breaker state
func (cb *CircuitBreaker) Execute(reqFunc func() error) error {
	generation, err := cb.beforeRequest()
	if err != nil {
		return err
	}

	defer func() {
		if e := recover(); e != nil {
			cb.afterRequest(generation, false)
			panic(e)
		}
	}()

	err = reqFunc()
	cb.afterRequest(generation, err == nil)
	return err
}

func (cb *CircuitBreaker) beforeRequest() (uint64, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)

	if state == StateOpen {
		return generation, ErrCircuitBreakerOpen
	} else if state == StateHalfOpen && cb.counts.Requests >= cb.maxRequests {
		return generation, ErrCircuitBreakerOpen
	}

	cb.counts.onRequest()
	return generation, nil
}

func (cb *CircuitBreaker) afterRequest(beforeGeneration uint64, success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()
	state, generation := cb.currentState(now)
	if generation != beforeGeneration {
		return
	}

	if success {
		cb.onSuccess(state, now)
	} else {
		cb.onFailure(state, now)
	}
}

func (cb *CircuitBreaker) onSuccess(state State, now time.Time) {
	cb.counts.onSuccess()

	if state == StateHalfOpen {
		if cb.counts.ConsecutiveSuccesses >= cb.maxRequests {
			cb.setState(StateClosed, now)
		}
	}
}

func (cb *CircuitBreaker) onFailure(state State, now time.Time) {
	cb.counts.onFailure()

	if state == StateClosed {
		if cb.counts.ConsecutiveFailures >= cb.failureThreshold {
			cb.setState(StateOpen, now)
		}
	} else if state == StateHalfOpen {
		cb.setState(StateOpen, now)
	}
}

func (cb *CircuitBreaker) currentState(now time.Time) (State, uint64) {
	switch cb.state {
	case StateClosed:
		if !cb.expiry.IsZero() && cb.expiry.Before(now) {
			cb.toNewGeneration(now)
		}
	case StateOpen:
		if cb.expiry.Before(now) {
			cb.setState(StateHalfOpen, now)
		}
	}
	return cb.state, cb.generation
}

func (cb *CircuitBreaker) setState(state State, now time.Time) {
	if cb.state == state {
		return
	}

	cb.state = state
	cb.toNewGeneration(now)
}

func (cb *CircuitBreaker) toNewGeneration(now time.Time) {
	cb.generation++
	cb.counts.clear()

	var zero time.Time
	switch cb.state {
	case StateClosed:
		if cb.interval > 0 {
			cb.expiry = now.Add(cb.interval)
		} else {
			cb.expiry = zero
		}
	case StateOpen:
		cb.expiry = now.Add(cb.timeout)
	default:
		cb.expiry = zero
	}
}

func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
