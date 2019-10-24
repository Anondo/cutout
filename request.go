package cutout

import (
	"bytes"
	"context"
	"errors"
	"net/http"
)

// Request ...
type Request struct {
	URL           string
	Method        string
	RequestBody   *bytes.Buffer
	Headers       map[string]string
	AllowedStatus []int
}

func (r *Request) isAllowedStatus(status int) bool {
	for _, sts := range r.AllowedStatus {
		if sts == status {
			return true
		}
	}
	return false
}

func (c *CircuitBreaker) makeRequest(r Request) (*Response, error) {

	req := &http.Request{}

	var err error

	if r.RequestBody != nil {
		req, err = http.NewRequest(r.Method, r.URL, r.RequestBody)
	} else {
		req, err = http.NewRequest(r.Method, r.URL, nil)
	}

	if err != nil {
		return nil, err
	}

	if r.Headers != nil {
		for key, value := range r.Headers {
			req.Header.Add(key, value)
		}
	}

	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), c.TimeOut)
	defer cancel()
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	bdy, err := getRespBodyString(resp.Body)

	if err != nil {
		return nil, err
	}

	if len(r.AllowedStatus) != 0 {
		if !r.isAllowedStatus(resp.StatusCode) {
			return nil, errors.New(bdy)
		}
	} else if resp.StatusCode >= 400 {
		return nil, errors.New(bdy)
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Body:       bdy,
	}, nil

}
