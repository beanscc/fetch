package fetch

import (
	"net/http"
)

type response struct {
	resp *http.Response
	body []byte
	err  error
}

// newErrResp return new resp with err
func newErrResp(e error) *response {
	return &response{
		err: e,
	}
}

// Resp 返回 http.Response
func (r *response) Resp() (*http.Response, error) {
	return r.resp, r.err
}

// Body 返回请求响应 body 消息体
func (r *response) Bytes() ([]byte, error) {
	return r.body, r.err
}

// Text 返回请求响应 body 消息体
func (r *response) Text() (string, error) {
	return string(r.body), r.err
}
