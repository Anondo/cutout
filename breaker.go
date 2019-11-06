package cutout

import (
	"net/http"
	"time"
)

// CircuitBreaker is the circuit breaker!!!
type CircuitBreaker struct {
	FailThreshold     int
	HealthCheckPeriod time.Duration
	events            chan string
	state             string
	lastFailed        *time.Time
	failCount         int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failThreshold int, healthCheckPeriod time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		FailThreshold:     failThreshold,
		HealthCheckPeriod: healthCheckPeriod,
	}
}

// Call calls an external service using the circuit breaker design
func (c *CircuitBreaker) Call(req Request, fallbackFuncs ...func() (*Response, error)) (*Response, error) {
	c.setState()

	var resp *Response
	var err error

	switch c.state {
	case ClosedState, HalfOpenState:
		resp, err = req.makeRequest()
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

// CallWithCustomRequest calls an external service using the circuit breaker design with a custom request function
func (c *CircuitBreaker) CallWithCustomRequest(req *http.Request, allowedStatus []int,
	fallbackFuncs ...func() (*Response, error)) (*Response, error) {
	c.setState()

	var resp *Response
	var err error

	switch c.state {
	case ClosedState, HalfOpenState:
		resp, err = makeCustomRequest(req, allowedStatus)
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
