package cutout

import "time"

type (

	// Failure holds the failure instance information
	Failure struct {
		Message       string    `json:"message"`
		OccurredAt    time.Time `json:"occurred_at"`
		TotalFailures int       `json:"total_failures"`
	}

	// RequestRecord holds the information of a request incident
	RequestRecord struct {
		Name        string    `json:"name"`
		Method      string    `json:"method"`
		StatusCode  int       `json:"status_code"`
		StatusText  string    `json:"status_text"`
		Message     string    `json:"message"`
		RequestedAt time.Time `json:"requested_at"`
	}

	// Analytics contains analytical informations regarding the circuit breaker
	Analytics struct {
		RequestSent    int             `json:"request_sent"`
		TotalFailures  int             `json:"total_failures"`
		FallbackCalls  int             `json:"fallback_calls"`
		Failures       []Failure       `json:"failures"`
		TotalCalls     int             `json:"total_calls"`
		SuccessRate    float64         `json:"success_rate"`
		FailureRate    float64         `json:"failure_rate"`
		RequestRecords []RequestRecord `json:"request_records"`
	}
)

// InitAnalytics initializes the analytics instance for the circu breaker to start analyzing
func (c *CircuitBreaker) InitAnalytics() {
	c.analytics = &Analytics{}
}

// GetAnalytics returns the analytics instance of the circuit breaker
func (c *CircuitBreaker) GetAnalytics() *Analytics {
	return c.analytics
}

func (c *CircuitBreaker) updateAnalyticsFailure(errMsg string) {
	if c.analytics != nil {
		c.analytics.TotalFailures++
		c.analytics.Failures = append(c.analytics.Failures, Failure{
			Message:       errMsg,
			OccurredAt:    time.Now(),
			TotalFailures: c.analytics.TotalFailures + 1,
		})
	}
}

func (c *CircuitBreaker) updateAnalyticsRequestAndResponse(url, method string, reqTime time.Time, resp *Response) {
	if c.analytics != nil {
		c.analytics.RequestSent++
		rr := RequestRecord{
			Name:        url,
			Method:      method,
			RequestedAt: reqTime,
		}
		if resp != nil {
			rr.StatusCode = resp.StatusCode
			rr.StatusText = resp.Status
			rr.Message = resp.BodyString
		}
		c.analytics.RequestRecords = append(c.analytics.RequestRecords, rr)
	}
}

func (c *CircuitBreaker) addAnalyticsFallbackCount() {
	if c.analytics != nil {
		c.analytics.FallbackCalls++
	}
}

func (c *CircuitBreaker) updateAnalyticsRates() {
	if c.analytics != nil {
		c.analytics.TotalCalls++
		c.analytics.SuccessRate = float64(c.analytics.RequestSent-c.analytics.TotalFailures) / float64(c.analytics.RequestSent) * 100
		c.analytics.FailureRate = 100 - c.analytics.SuccessRate
	}
}
