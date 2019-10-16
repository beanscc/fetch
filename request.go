package fetch

import (
	"net/http"
	"net/url"

	"github.com/beanscc/fetch/body"
)

var allowedBodyMethods = map[string]bool{
	http.MethodPost: true,
	http.MethodPut:  true,
}

type request struct {
	url    *url.URL          // req url
	method string            // req method
	params map[string]string // req query 参数
	header http.Header       // req header
	body   body.Body         // req body
}

func newRequest() *request {
	return &request{
		url:    new(url.URL),
		params: make(map[string]string),
		header: make(http.Header),
		body:   nil,
	}
}

// isAllowedBody 检查method是否允许设置body
func isAllowedBody(method string) bool {
	return allowedBodyMethods[method]
}
