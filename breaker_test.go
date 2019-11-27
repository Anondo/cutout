package cutout

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const (
	numberOfCalls       = 5
	mockServerURL       = "127.0.0.1:9001"
	mockServerURLCustom = "127.0.0.1:9002"
)

type (
	testEventInfo struct {
		currentState *string
		failed       bool
	}
)

func TestCircuitBreakerCall(t *testing.T) {

	cb := NewCircuitBreaker(2, 5*time.Second)

	event := make(chan string, 1)
	cb.InitEvent(event)
	tei := &testEventInfo{}

	go func() {
		for {
			switch <-event {
			case StateChangeEvent:
				state := cb.State()
				tei.currentState = &state
			case FailureEvent:
				tei.failed = true
			}
		}
	}()

	ts := testServer(mockServerURL)

	cache := `{"name":"Mr. Test","age":69,"cgpa":4}`

	defer ts.Close()

	handler := testCallHandler(ts.URL, cb, cache)

	if err := startIntegrationTest(t, mockServerURL, ts, cb, handler, tei); err != nil {
		t.Errorf(err.Error())
	}

}

func TestCircuitBreakerCallWithCustomRequest(t *testing.T) {

	cb := NewCircuitBreaker(2, 5*time.Second)

	event := make(chan string, 1)
	cb.InitEvent(event)
	tei := &testEventInfo{}

	go func() {
		for {
			switch <-event {
			case StateChangeEvent:
				state := cb.State()
				tei.currentState = &state
			case FailureEvent:
				tei.failed = true
			}
		}
	}()

	ts := testServer(mockServerURLCustom)

	cache := `{"name":"Mr. Test","age":69,"cgpa":4}`

	defer ts.Close()

	handler := testCallWithCustomRequestHandler(ts.URL, cb, cache)

	if err := startIntegrationTest(t, mockServerURLCustom, ts, cb, handler, tei); err != nil {
		t.Errorf(err.Error())
	}

}

func startIntegrationTest(t *testing.T, url string, ts *httptest.Server, cb *CircuitBreaker,
	handler func(w http.ResponseWriter, r *http.Request), tei *testEventInfo) error {
	// #########################
	// # Checking closed state #
	// #########################

	t.Log("Checking closed state...")
	for i := 0; i < numberOfCalls; i++ {
		t.Logf("Call:%d, state:%s, status:%d, fail count:%d\n", i+1, ClosedState, http.StatusOK, 0)
		resp := testCall(handler)
		if err := checkErrors(http.StatusOK, resp.StatusCode, ClosedState,
			cb.State(), 0, cb.FailCount()); err != nil {
			return err
		}
		time.Sleep(50 * time.Millisecond) //waiting for the event to fire, just for this test
		if err := checkEvents(tei, ClosedState, false); err != nil {
			return err
		}
	}

	t.Log("PASSED")

	ts.Close() // closing the third-party service

	// ###########################
	// # Checking fail threshold #
	// ###########################

	t.Log("Checking fail threshold...")
	for i := 0; i < cb.FailThreshold; i++ {
		t.Logf("Call:%d, state:%s, status:%d, fail count:%d\n", i+1, ClosedState, http.StatusInternalServerError, i+1)
		resp := testCall(handler)
		if err := checkErrors(http.StatusInternalServerError, resp.StatusCode,
			ClosedState, cb.State(), i+1, cb.FailCount()); err != nil {
			return err
		}
		time.Sleep(50 * time.Millisecond)
		if err := checkEvents(tei, ClosedState, true); err != nil {
			return err
		}
	}

	t.Log("PASSED")

	// #######################
	// # Checking open state #
	// #######################

	t.Log("Checking open state...")
	for i := 0; i < numberOfCalls; i++ {
		t.Logf("Call:%d, state:%s, status:%d, fail count:%d\n", i+1, OpenState, http.StatusOK, cb.FailThreshold)
		resp := testCall(handler)
		if err := checkErrors(http.StatusOK, resp.StatusCode, OpenState,
			cb.State(), cb.FailThreshold, cb.FailCount()); err != nil {
			return err
		}
		time.Sleep(50 * time.Millisecond)
		if err := checkEvents(tei, OpenState, false); err != nil {
			return err
		}
	}

	t.Log("PASSED")

	t.Logf("Waiting %v for half open state...\n", cb.HealthCheckPeriod)
	time.Sleep(cb.HealthCheckPeriod) // waiting for the health check period for half open state

	// ############################
	// # Checking half open state #
	// ############################
	t.Log("Checking half open state...")
	t.Logf("state:%s, status:%d, fail count:%d\n", HalfOpenState, http.StatusInternalServerError, cb.FailThreshold+1)
	respStatusCode := testCall(handler).StatusCode
	if err := checkErrors(http.StatusInternalServerError, respStatusCode,
		HalfOpenState, cb.State(), cb.FailThreshold+1, cb.FailCount()); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	if err := checkEvents(tei, HalfOpenState, true); err != nil {
		return err
	}

	t.Log("PASSED")

	// #####################################################
	// # Checking open state after half open state failure #
	// #####################################################
	t.Log("Checking open state after the half open state...")
	for i := 0; i < numberOfCalls; i++ {
		t.Logf("Call:%d, state:%s, status:%d, fail count:%d\n", i+1, OpenState, http.StatusOK, cb.FailThreshold+1)
		resp := testCall(handler)
		if err := checkErrors(http.StatusOK, resp.StatusCode, OpenState,
			cb.State(), cb.FailThreshold+1, cb.FailCount()); err != nil {
			return err
		}
		time.Sleep(50 * time.Millisecond)
		if err := checkEvents(tei, OpenState, false); err != nil {
			return err
		}
	}

	t.Log("PASSED")

	ts = testServer(url) // restarting the third-party service

	t.Logf("Waiting %v for half open state...\n", cb.HealthCheckPeriod)
	time.Sleep(cb.HealthCheckPeriod)

	// #######################################
	// # Checking the second half open state #
	// #######################################
	t.Log("Checking the second half open state...")
	t.Logf("state:%s, status:%d, fail count:%d\n", HalfOpenState, http.StatusOK, 0)
	respStatusCode = testCall(handler).StatusCode
	if err := checkErrors(http.StatusOK, respStatusCode, HalfOpenState,
		cb.State(), 0, cb.FailCount()); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	if err := checkEvents(tei, HalfOpenState, false); err != nil {
		return err
	}

	t.Log("PASSED")

	// #######################################
	// # Finally Checking closed state again #
	// #######################################
	t.Log("Checking closed state after half open, as the server came live")
	for i := 0; i < numberOfCalls; i++ {
		t.Logf("Call:%d, state:%s, status:%d, fail count:%d\n", i+1, ClosedState, http.StatusOK, 0)
		resp := testCall(handler)
		if err := checkErrors(http.StatusOK, resp.StatusCode, ClosedState,
			cb.State(), 0, cb.FailCount()); err != nil {
			return err
		}
		time.Sleep(50 * time.Millisecond)
		if err := checkEvents(tei, ClosedState, false); err != nil {
			return err
		}
	}

	t.Log("PASSED")

	return nil
}

func testCall(handler func(w http.ResponseWriter, r *http.Request)) *http.Response {
	req := httptest.NewRequest(http.MethodGet, "http://cutout.hehe", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()

	// body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println(string(body))
	return resp

}

func checkErrors(wantStatus, gotStatus int,
	wantState, gotState string, wantFailCount, gotFailCount int) error {
	if wantStatus != gotStatus {
		return fmt.Errorf("Incorrent status code received, wanted %d got %d", wantStatus, gotStatus)
	}

	if wantState != gotState {
		return fmt.Errorf("Incorrect state of the circuit received, wanted %s got %s", wantState, gotState)
	}

	if wantFailCount != gotFailCount {
		return fmt.Errorf("Fail count should have been %d, got %d", wantFailCount, gotFailCount)
	}

	return nil
}

func testServer(url string) *httptest.Server {
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "application/json")

		stdnt := struct {
			Name string  `json:"name"`
			Age  int     `json:"age"`
			CGPA float64 `json:"cgpa"`
		}{
			Name: "Tony",
			Age:  69,
			CGPA: 4.00,
		}

		bb, _ := json.Marshal(stdnt)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, string(bb))
	}))

	lstnr, err := net.Listen("tcp", url)

	if err != nil {
		log.Fatal(err.Error())
	}
	ts.Listener = lstnr

	ts.Start()

	return ts
}

func testCallHandler(url string, cb *CircuitBreaker, cache string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "application/json")
		req := Request{
			URL:           url,
			AllowedStatus: []int{http.StatusOK},
			Method:        http.MethodGet,
			TimeOut:       2 * time.Second,
		}
		resp, err := cb.Call(&req, func() (*Response, error) {
			return &Response{
				BodyString: cache,
				Response: &http.Response{
					StatusCode: http.StatusOK,
				},
			}, nil
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, `{"message": "Something went wrong","error":`+fmt.Sprintf(`"%s"`, err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, resp.BodyString)
	}
}

func testCallWithCustomRequestHandler(url string, cb *CircuitBreaker, cache string) func(http.ResponseWriter,
	*http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "application/json")

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, `{"message": "Something went wrong","error":`+fmt.Sprintf(`"%s"`, err.Error()))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		req = req.WithContext(ctx)
		resp, err := cb.CallWithCustomRequest(req, []int{http.StatusOK}, func() (*Response, error) {
			return &Response{
				BodyString: cache,
				Response: &http.Response{
					StatusCode: http.StatusOK,
				},
			}, nil
		})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, `{"message": "Something went wrong","error":`+fmt.Sprintf(`"%s"`, err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, resp.BodyString)
	}
}

func checkEvents(tei *testEventInfo, wantState string, wantFail bool) error {
	if tei.currentState == nil {
		return fmt.Errorf("Invalid state received, wanted:%s, got:%v", wantState, tei.currentState)
	}
	if *tei.currentState != wantState {
		return fmt.Errorf("Unexpected state received from event, wanted %s got %s", wantState, *tei.currentState)
	}

	if tei.failed != wantFail {
		return fmt.Errorf("Unexpected failure event, wanted failure to be:%v, got:%v", wantFail, tei.failed)
	}

	tei.failed = false

	return nil
}
