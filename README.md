# Cutout

[![Build Status](https://travis-ci.org/Anondo/cutout.svg?branch=master)](https://travis-ci.org/Anondo/cutout)
[![License](https://img.shields.io/dub/l/vibe-d.svg)](https://github.com/Anondo/cutout/blob/master/LICENSE)
[![Project status](https://img.shields.io/badge/version-1.0.0-green.svg)](https://github.com/Anondo/cutout/releases)
[![GoDoc](https://godoc.org/github.com/Anondo/cutout?status.svg)](https://godoc.org/github.com/Anondo/cutout)


Package cutout implements the circuit breaker design pattern(see: https://martinfowler.com/bliki/CircuitBreaker.html)
for calling third party api services.

Cutout comes with features like:

1. Multilevel fallback functions(in case even the fallback fails)
1. Custom BackOff function on the request level for generating backoff timeout logics
1. Event channel to capture events like State change or failure detection
1. Get analytical data on the circuit breaker

### Installing
```console
go get -u github.com/Anondo/cutout

```

### Usage

**Import The Package**

```go
import "github.com/Anondo/cutout"

```

**Initiate the circuit breaker and request instances**

```go
 var (
	 cb = &cutout.CircuitBreaker{
	  	FailThreshold:     100,
	  	HealthCheckPeriod: 15 * time.Second,
	 }
	 req = cutout.Request{
	  	URL:           "http://localhost:9090",
	  	AllowedStatus: []int{http.StatusOK},
	  	Method:        http.MethodPost,
	  	TimeOut:       2 * time.Second,
	  	RequestBody:   bytes.NewBuffer([]byte(`{"name":"abcd"}`)),
	  	BackOff: func(t time.Duration) time.Duration {
		    return time.Duration(int(t/time.Second)*5) * time.Second
	      },
	 }

 )

```

**Prepare a fallback function**
```go

func() theFallbackFunc(*cutout.Response, error) {
  // some fallback codes here...

  return &cutout.Response{
    BodyString: cachedResponse,
  }, nil

}

```

**Call a third party service from your handler**

```go

func thehandler(w http.ResponseWriter, r *http.Request) {

	resp, err := cb.Call(&req, theFallbackFunc)

	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, resp.BodyString)
}

```

For more details, see the [docs](https://godoc.org/github.com/Anondo/cutout) and [examples](examples/).


### Contributing

See the contributions guide [here](CONTRIBUTING.md).

### License

Cutout is licensed under the [MIT License](LICENSE).
