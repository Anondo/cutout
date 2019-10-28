package cutout

import (
	"time"
)

// states of the circuit breaker
const (
	ClosedState   = "CLOSED"
	OpenState     = "OPEN"
	HalfOpenState = "HALF_OPEN"
)

// determine the current the state of the circuit
func (c *CircuitBreaker) setState() {
	prevState := c.state

	if c.failCount >= c.FailThreshold {
		if time.Now().Sub(*c.lastFailed) > c.HealthCheckPeriod { //the time for health check has arrived
			c.state = HalfOpenState
		} else {
			c.state = OpenState
		}
	} else {
		c.state = ClosedState //everything is good
	}

	if c.state != prevState {
		c.fireEvent(StateChangeEvent)
	}

}

// reset the circuit to its initial state
func (c *CircuitBreaker) resetCircuit() {
	c.failCount = 0
	c.lastFailed = nil
}

// State returns the current satte of the circuit
func (c *CircuitBreaker) State() string {
	return c.state
}

// FailCount returns the count of failure
func (c *CircuitBreaker) FailCount() int {
	return c.failCount
}

// LastFailed returns the time object of the last failure
func (c *CircuitBreaker) LastFailed() *time.Time {
	return c.lastFailed
}

func (c *CircuitBreaker) updateFailData() {
	c.fireEvent(FailureEvent)
	now := time.Now()
	c.lastFailed = &now
	c.failCount++
}
