//Package cutout implements the circuit breaker design pattern(see: https://martinfowler.com/bliki/CircuitBreaker.html)
//for calling third party api services.
//
// Cutout comes with features like:
//
// 1. Multilevel fallback functions(in case even the fallback fails)
//
// 2. Custom BackOff function on the request level for generating backoff timeout logics
//
// 3. Event channel to capture events like State change or failure detection
//
//Here is a basic example:
//  package main
//
//  import (
// 	 "bytes"
// 	 "fmt"
// 	 "log"
// 	 "net/http"
// 	 "time"
//
// 	 "github.com/Anondo/cutout"
//  )
//
//  var (
// 	 cb = &cutout.CircuitBreaker{
// 	  	FailThreshold:     100,
// 	 	HealthCheckPeriod: 15 * time.Second,
// 	 }
// 	 req = cutout.Request{
// 	  	URL:           "http://localhost:9090",
// 	  	AllowedStatus: []int{http.StatusOK},
// 	  	Method:        http.MethodPost,
// 	  	TimeOut:       2 * time.Second,
// 	  	RequestBody:   bytes.NewBuffer([]byte(`{"name":"abcd"}`)),
// 		BackOff: func(t time.Duration) time.Duration {
// 		  return time.Duration(int(t/time.Second)*5) * time.Second
// 	        },
// 	 }
//
// 	 cache = `{"name":"Mr. Test","age":69,"cgpa":4}`
//  )
//
//  func thehandler(w http.ResponseWriter, r *http.Request) {
//
// 	 resp, err := cb.Call(&req, func() (*cutout.Response, error) {
// 	 	 return &cutout.Response{
// 			 BodyString: cache,
// 		 }, nil
// 	 })
//
// 	 if err != nil {
// 		 fmt.Fprintf(w, err.Error())
// 		 return
// 	 }
//
// 	 w.WriteHeader(http.StatusOK)
// 	 fmt.Fprintf(w, resp.BodyString)
//  }
//
//  func main() {
//
// 	 http.HandleFunc("/", thehandler)
//
// 	 log.Println("Service A is running on http://localhost:8080 ...")
//
// 	 if err := http.ListenAndServe(":8080", nil); err != nil {
// 		 log.Fatal(err.Error())
// 	 }
//  }
//
// See https://github.com/Anondo/cutout/examples.
package cutout
