package cutout

import (
	"time"
)

// CircuitBreaker is the circuit breaker!!!
type CircuitBreaker struct {
	FailThreshold     int
	TimeOut           time.Duration
	HealthCheckPeriod time.Duration
	state             string
	lastFailed        *time.Time
	failCount         int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failThreshold int, timeout, healthCheckPeriod time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		FailThreshold:     failThreshold,
		TimeOut:           timeout,
		HealthCheckPeriod: healthCheckPeriod,
	}
}

// Call calls an external service
func (c *CircuitBreaker) Call(req Request, fallbackFuncs ...func() (*Response, error)) (*Response, error) {
	c.setState()

	var resp *Response
	var err error

	switch c.state {
	case ClosedState, HalfOpenState:
		resp, err = c.makeRequest(req)
		if err != nil {
			c.updateFailData()
		} else {
			c.resetCircuit()
		}
	case OpenState:
		resp, err = executeFallbacks(fallbackFuncs)
		if err != nil {
			return resp, err
		}
	}

	return resp, err
}

func (c *CircuitBreaker) updateFailData() {
	now := time.Now()
	c.lastFailed = &now
	c.failCount++
}
