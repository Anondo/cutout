package main

import (
	"net/http"
	"time"

	"github.com/Anondo/cutout"
)

var (
	cb = &cutout.CircuitBreaker{
		HealthCheckPeriod: 30 * time.Second,
		FailThreshold:     10,
	}
	cache = `{"message":"PING!"}`
)

func fallBackFunc() (*cutout.Response, error) {
	return &cutout.Response{
		BodyString: cache,
		Response: &http.Response{
			StatusCode: http.StatusOK,
		},
	}, nil
}

func getPingFromService() (string, error) {
	pingRequest := &cutout.Request{
		URL:           "http://thisisanexampleurl.com",
		Method:        http.MethodGet,
		RequestBody:   nil,
		AllowedStatus: []int{http.StatusOK},
		TimeOut:       10 * time.Second,
	}

	resp, err := cb.Call(pingRequest, fallBackFunc)

	if err != nil {
		return "", err
	}

	return resp.BodyString, nil
}
