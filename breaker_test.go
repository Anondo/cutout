package cutout

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var (
	numberOfCalls = 5
	url           = "127.0.0.1:9001"
)

func TestResponseRecorder(t *testing.T) {

	cb := &CircuitBreaker{
		FailThreshold:     2,
		HealthCheckPeriod: 5 * time.Second,
	}

	ts := testServer()

	cache := `{"name":"Mr. Test","age":69,"cgpa":4}`

	defer ts.Close()

	handler := testHandler(ts, cb, cache)

	// #########################
	// # Checking closed state #
	// #########################

	t.Log("Checking closed state...")
	for i := 0; i < numberOfCalls; i++ {
		t.Logf("Call:%d, state:%s, status:%d, fail count:%d\n", i+1, ClosedState, http.StatusOK, 0)
		if err := checkForErrors(testCall(handler), cb, http.StatusOK, ClosedState, 0); err != nil {
			t.Errorf(err.Error())
			return
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
		if err := checkForErrors(testCall(handler), cb, http.StatusInternalServerError, ClosedState, i+1); err != nil {
			t.Errorf(err.Error())
			return
		}
	}

	t.Log("PASSED")

	// #######################
	// # Checking open state #
	// #######################

	t.Log("Checking open state...")
	for i := 0; i < numberOfCalls; i++ {
		t.Logf("Call:%d, state:%s, status:%d, fail count:%d\n", i+1, OpenState, http.StatusOK, cb.FailThreshold)
		if err := checkForErrors(testCall(handler), cb, http.StatusOK, OpenState, cb.FailThreshold); err != nil {
			t.Errorf(err.Error())
			return
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
	if err := checkForErrors(testCall(handler), cb, http.StatusInternalServerError, HalfOpenState, cb.FailThreshold+1); err != nil {
		t.Errorf(err.Error())
		return
	}

	t.Log("PASSED")

	// #####################################################
	// # Checking open state after half open state failure #
	// #####################################################
	t.Log("Checking open state after the half open state...")
	for i := 0; i < numberOfCalls; i++ {
		t.Logf("Call:%d, state:%s, status:%d, fail count:%d\n", i+1, OpenState, http.StatusOK, cb.FailThreshold+1)
		if err := checkForErrors(testCall(handler), cb, http.StatusOK, OpenState, cb.FailThreshold+1); err != nil {
			t.Errorf(err.Error())
			return
		}
	}

	t.Log("PASSED")

	ts = testServer() // restarting the third-party service

	t.Logf("Waiting %v for half open state...\n", cb.HealthCheckPeriod)
	time.Sleep(cb.HealthCheckPeriod)

	// #######################################
	// # Checking the second half open state #
	// #######################################
	t.Log("Checking the second half open state...")
	t.Logf("state:%s, status:%d, fail count:%d\n", HalfOpenState, http.StatusOK, 0)
	if err := checkForErrors(testCall(handler), cb, http.StatusOK, HalfOpenState, 0); err != nil {
		t.Errorf(err.Error())
		return
	}

	t.Log("PASSED")

	// #######################################
	// # Finally Checking closed state again #
	// #######################################
	t.Log("Checking closed state after half open, as the server came live")
	for i := 0; i < numberOfCalls; i++ {
		t.Logf("Call:%d, state:%s, status:%d, fail count:%d\n", i+1, ClosedState, http.StatusOK, 0)
		if err := checkForErrors(testCall(handler), cb, http.StatusOK, ClosedState, 0); err != nil {
			t.Errorf(err.Error())
			return
		}
	}

	t.Log("PASSED")

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

func checkForErrors(resp *http.Response, cb *CircuitBreaker, status int,
	state string, failCount int) error {
	if resp.StatusCode != status {
		return fmt.Errorf("Incorrent status code received, wanted %d got %d", status, resp.StatusCode)
	}

	if cb.State() != state {
		return fmt.Errorf("Incorrect state of the circuit received, wanted %s got %s", state, cb.State())
	}

	if cb.FailCount() != failCount {
		return fmt.Errorf("Fail count should have been %d, got %d", failCount, cb.FailCount())
	}

	return nil
}

func testServer() *httptest.Server {
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

	lstnr, _ := net.Listen("tcp", url)
	ts.Listener = lstnr

	ts.Start()

	return ts
}

func testHandler(ts *httptest.Server, cb *CircuitBreaker, cache string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-type", "application/json")
		req := Request{
			URL:           ts.URL,
			AllowedStatus: []int{http.StatusOK},
			Method:        http.MethodGet,
			TimeOut:       2 * time.Second,
		}
		resp, err := cb.Call(req, func() (*Response, error) {
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
