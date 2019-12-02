package cutout

// Events
const (
	StateChangeEvent = "STATE_CHANGE"
	FailureEvent     = "FAILURE"
)

// InitEvent initializes the circuit breaker events
// NOTE: the parameter must be a buffered channel
//
// Parameters:
//
// 1. chan string ------> the event channel, must be a buffered channel so that the circuit doesn't get blocked out
//
// Example:
//
//  events := make(chan string, 2)
//
//  cb.InitEvent(events)
//
//  go func() {
// 	 log.Println("Waiting for events from circuit breaker...")
// 	 for {
// 		 switch <-events {
// 		 case cutout.StateChangeEvent:
// 			 log.Println("Status change occured")
// 			 log.Println("Current state:", cb.State())
// 		 case cutout.FailureEvent:
// 			 log.Println("Failure occured")
// 		 }
// 	 }
// 	 log.Println("Done with the waiting")
//  }()
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
