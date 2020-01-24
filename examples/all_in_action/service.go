package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Anondo/cutout"
)

type pingResponse struct {
	Message       string `json:"message"`
	SuccessRate   string `json:"success_rate"`
	FailureRate   string `json:"failure_rate"`
	RequestSent   int    `json:"request_sent"`
	TotalFailures int    `json:"total_failures"`
}

var (
	cb = &cutout.CircuitBreaker{
		HealthCheckPeriod: 30 * time.Second,
		FailThreshold:     10,
	}
	events = make(chan string, 1)
	cache  = `{"message":"PING!"}`
)

func init() {
	cb.InitAnalytics()
	cb.InitEvent(events)
	go func() {
		for {
			switch <-events {
			case cutout.StateChangeEvent:
				log.Println("State change event occured!")
				log.Println("Current state: ", cb.State())
			case cutout.FailureEvent:
				log.Println("Failure occured!")
				log.Println("Failure time: ", cb.LastFailed())
			}
		}
	}()
}

func fallBackFunc() (*cutout.Response, error) {
	return &cutout.Response{
		BodyString: cache,
		Response: &http.Response{
			StatusCode: http.StatusOK,
		},
	}, nil
}

func getPingFromService() (*pingResponse, error) {
	pingRequest := &cutout.Request{
		URL:           "http://thisisanexampleurl.com",
		Method:        http.MethodGet,
		RequestBody:   nil,
		AllowedStatus: []int{http.StatusOK},
		TimeOut:       10 * time.Second,
		BackOff: func(t time.Duration) time.Duration {
			return time.Duration(int(t/time.Second)*5) * time.Second
		},
	}

	resp, err := cb.Call(pingRequest, fallBackFunc)

	if err != nil {
		return nil, err
	}

	pr := pingResponse{}

	if err := json.Unmarshal([]byte(resp.BodyString), &pr); err != nil {
		return nil, err
	}

	analytics := cb.GetAnalytics()

	pr.SuccessRate = fmt.Sprintf("%.2f%%", analytics.SuccessRate)
	pr.FailureRate = fmt.Sprintf("%.2f%%", analytics.FailureRate)
	pr.RequestSent = analytics.RequestSent
	pr.TotalFailures = analytics.TotalFailures

	return &pr, nil
}
