package fetch

import (
	"io"
	"net/http"
)

type request struct {
	*http.Request
	body io.Reader
}

func newRequest() *request {
	return &request{
		Request: &http.Request{
			Header: make(http.Header),
		},
		body: nil,
	}
}
