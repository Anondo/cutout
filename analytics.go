package cutout

import "time"

type (

	// Failure holds the failure instance information
	Failure struct {
		Message       string    `json:"message"`
		OccurredAt    time.Time `json:"occurred_at"`
		TotalFailures int       `json:"total_failures"`
	}

	// Analytics contains analytical informations regarding the circuit breaker
	Analytics struct {
		RequestSent   int       `json:"request_sent"`
		TotalFailures int       `json:"total_failures"`
		FallbackCalls int       `json:"fallback_calls"`
		Failures      []Failure `json:"failures"`
		TotalCalls    int       `json:"total_calls"`
		SuccessRate   float64   `json:"success_rate"`
		FailureRate   float64   `json:"faliure_rate"`
	}
)

// NewAnalytics creates a new and empty analytics instance
func NewAnalytics() *Analytics {
	return &Analytics{}
}

// InitAnalytics initializes the analytics instance for the circu breaker to start analyzing
func (c *CircuitBreaker) InitAnalytics(anltcs *Analytics) {
	c.analytics = anltcs
}

// GetAnalytics returns the analytics instance of the circuit breaker
func (c *CircuitBreaker) GetAnalytics() *Analytics {
	return c.analytics
}
