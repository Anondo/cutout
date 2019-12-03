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
	analytics         *Analytics
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(failThreshold int, healthCheckPeriod time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		FailThreshold:     failThreshold,
		HealthCheckPeriod: healthCheckPeriod,
	}
}

// Call calls an external service using the circuit breaker design
//
// Parameters:
//
// 1. *cutout.Request -------> The request object
//
// 2. ...func()(*Response , error) -----> one or many fallback functions which must return a *cutout.Response & error instance
//
// Example:
//
//  resp, err := cb.Call(&req, func() (*cutout.Response, error) {
// 	 return &cutout.Response{
// 	 	 BodyString: cache,
//   }, nil
//  })
func (c *CircuitBreaker) Call(req *Request, fallbackFuncs ...func() (*Response, error)) (*Response, error) {
	c.setState()

	var resp *Response
	var err error

	switch c.state {
	case ClosedState, HalfOpenState:
		resp, err = req.makeRequest()
		if err != nil {
			c.updateFailData()
			if c.analytics != nil {
				c.analytics.Failures = append(c.analytics.Failures, Failure{
					Message:       err.Error(),
					OccurredAt:    time.Now(),
					TotalFailures: c.analytics.TotalFailures + 1,
				})
			}
		} else {
			if c.analytics != nil {
				c.analytics.RequestSent++
			}
			c.resetCircuit()
		}
	case OpenState:
		resp, err = executeFallbacks(fallbackFuncs)
		if err != nil {
			return resp, err
		}
		if c.analytics != nil {
			c.analytics.FallbackCalls++
		}
	}

	if c.analytics != nil {
		c.analytics.TotalCalls++
		c.analytics.SuccessRate = float64(c.analytics.RequestSent+c.analytics.FallbackCalls) / float64(c.analytics.TotalCalls) * 100
		c.analytics.FailureRate = 100 - c.analytics.SuccessRate
	}

	return resp, err
}

// CallWithCustomRequest calls an external service using the circuit breaker design with a custom request function
//
// Parameters:
//
// 1. *http.Request -------> The request object of the built-in http package
//
// 2. []int -------> Allowed http status codes, which wont be counted as failures
//
// 3. ...func()(*Response , error) -----> one or many fallback functions which must return a *cutout.Response & error instance
//
// Example:
//
//  req, err := http.NewRequest(http.MethodGet, url, nil)
//  if err != nil {
// 	return err
//  }
//
//  ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
//  defer cancel()
//  req = req.WithContext(ctx)
//  resp, err := cb.CallWithCustomRequest(req, []int{http.StatusOK}, func() (*Response, error) {
// 	 return &Response{
// 		 BodyString: cache,
// 		 Response: &http.Response{
// 			 StatusCode: http.StatusOK,
// 		 },
// 	 }, nil
//  })
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
