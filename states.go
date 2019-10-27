package cutout

import "time"

// states of the circuit breaker
const (
	ClosedState   = "CLOSED"
	OpenState     = "OPEN"
	HalfOpenState = "HALF_OPEN"
)

// determine the current the state of the circuit
func (c *CircuitBreaker) setState() {
	if c.failCount >= c.FailThreshold {
		if time.Now().Sub(*c.lastFailed) > c.HealthCheckPeriod { //the time for health check has arrived
			c.state = HalfOpenState
		} else {
			c.state = OpenState
		}
	} else {
		c.state = ClosedState //everything is good
	}

}

// reset the circuit to its initial state
func (c *CircuitBreaker) resetCircuit() {
	c.state = ClosedState
	c.failCount = 0
	c.lastFailed = nil
}

// State returns the current satte of the circuit
func (c *CircuitBreaker) State() string {
	return c.state
}
