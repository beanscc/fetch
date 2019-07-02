package fetch

import (
	"net/http"
	"net/url"

	"github.com/beanscc/fetch/body"
)

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

// 检查method是否允许设置body
func allowBody(method string) bool {
	allowed := []string{
		http.MethodPost,
		http.MethodPut,
	}

	for _, v := range allowed {
		if v == method {
			return true
		}
	}

	return false
}
