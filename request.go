package cutout

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"time"
)

// Request represents the data needed to make http requests
type Request struct {
	URL           string
	Method        string
	RequestBody   *bytes.Buffer
	Headers       map[string]string
	AllowedStatus []int
	TimeOut       time.Duration
}

// NewRequest is the factory function for requests i.e, creates a new request
func NewRequest(url, method string, headers map[string]string,
	requestBody *bytes.Buffer, allowedStatus []int, timeout time.Duration) Request {
	return Request{
		URL:           url,
		Method:        method,
		RequestBody:   requestBody,
		Headers:       headers,
		AllowedStatus: allowedStatus,
		TimeOut:       timeout,
	}
}

func (r *Request) isAllowedStatus(status int) bool {
	for _, sts := range r.AllowedStatus {
		if sts == status {
			return true
		}
	}
	return false
}

func (r *Request) makeRequest() (*Response, error) {

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
	ctx, cancel := context.WithTimeout(context.Background(), r.TimeOut)
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

	finalResponse := &Response{resp, bdy}

	if len(r.AllowedStatus) != 0 {
		if !r.isAllowedStatus(resp.StatusCode) {
			return finalResponse, errors.New(bdy)
		}
	} else if resp.StatusCode >= 400 {
		return finalResponse, errors.New(bdy)
	}

	return finalResponse, nil

}
