package cutout

// Events
const (
	StateChangeEvent = "STATE_CHANGE"
	FailureEvent     = "FAILURE"
)

// InitEvent initializes the circuit breaker events
// NOTE: the parameter must be a buffered channel
func (c *CircuitBreaker) InitEvent(e chan string) {
	if cap(e) > 0 {
		c.events = e
	}
}

func (c *CircuitBreaker) fireEvent(event string) {
	if cap(c.events) > 0 {
		c.events <- event
	}
}
