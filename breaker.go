package cb

import (
	"bytes"
	"time"
)

// CircuitBreaker is the circuit breaker!!!
type CircuitBreaker struct {
	FailThreshold     int
	TimeOut           time.Duration
	HealthCheckPeriod time.Duration
	state             string
	lastFailed        time.Time
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
func (c *CircuitBreaker) Call(url, method string, respBody bytes.Buffer, fallbackFuncs ...func() error) (Response, error) {
	c.setState()

	return Response{}, nil
}
