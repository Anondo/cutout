package cutout

import (
	"io"
	"io/ioutil"
)

// Response represents the response data from an http service
type Response struct {
	StatusCode int
	Body       string
}

func getRespBodyString(bdy io.Reader) (string, error) {
	bb, err := ioutil.ReadAll(bdy)

	if err != nil {
		return "", err
	}

	return string(bb), nil
}
