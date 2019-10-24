package cutout

import "time"

// states of the circuit breaker
const (
	ClosedState   = "CLOSED"
	OpenState     = "OPEN"
	HalfOpenState = "HALF_OPEN"
)

func (c *CircuitBreaker) setState() {
	if c.failCount >= c.FailThreshold {
		if time.Now().Sub(*c.lastFailed) > c.HealthCheckPeriod {
			c.state = HalfOpenState
		} else {
			c.state = OpenState
		}
	} else {
		c.state = ClosedState
	}

}

func (c *CircuitBreaker) resetCircuit() {
	c.state = ClosedState
	c.failCount = 0
	c.lastFailed = nil
}

// State returns the current satte of the circuit
func (c *CircuitBreaker) State() string {
	return c.state
}
