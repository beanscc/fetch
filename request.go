package fetch

import (
	"net/http"
	"net/url"
)

type request struct {
	url            *url.URL          // req url
	method         string            // req method
	queryParameter map[string]string // req query 参数
	header         http.Header       // req header
	body           Body              // req body
}

func newRequest() *request {
	return &request{
		url:            new(url.URL),
		queryParameter: make(map[string]string),
		header:         make(http.Header),
		body:           nil,
	}
}
