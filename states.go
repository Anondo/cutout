package cb

import "time"

// states of the circuit breaker
const (
	ClosedState   = "CLOSED"
	OpenState     = "OPEN"
	HalfOpenState = "HALF_OPEN"
)

func (c *CircuitBreaker) setState() {
	if c.failCount > c.FailThreshold {
		if time.Duration(time.Now().Unix()-c.lastFailed.Unix()) > c.HealthCheckPeriod*time.Millisecond {
			c.state = HalfOpenState
		} else {
			c.state = OpenState
		}
	} else {
		c.state = HalfOpenState
	}
}
