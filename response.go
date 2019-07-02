package fetch

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/beanscc/fetch/binding"
)

type response struct {
	resp *http.Response
	body []byte
	err  error
	bind map[string]binding.Binding
}

// newErrResp return new resp with err
func newErrResp(e error) *response {
	return &response{
		err: e,
	}
}

// Bind bind http.Response
func (r *response) Bind(bindType string, v interface{}) error {
	if r.err != nil {
		return r.err
	}

	if r.resp == nil {
		return errors.New("nil http.Response")
	}

	if b, ok := r.bind[bindType]; ok {
		return b.Bind(r.resp, r.body, v)
	}

	return fmt.Errorf("unknown bind type:%v", bindType)
}

// BindJSON json 格式解析 body 数据，并绑定至 obj 对象上，obj 必须是指针对象
func (r *response) BindJSON(v interface{}) error {
	return r.Bind("json", v)
}

// BindXML xml 格式解析 body 数据，并绑定至 v 对象上，v 必须是指针对象
func (r *response) BindXML(v interface{}) error {
	return r.Bind("xml", v)
}

// Resp 返回 http.Response
func (r *response) Resp() (*http.Response, error) {
	return r.resp, r.err
}

// Body 返回请求响应 body 消息体
func (r *response) Body() ([]byte, error) {
	return r.body, r.err
}

// 返回错误，可能是请求参数的错误，也可能是请求响应的错误
func (r *response) Error() error {
	return r.err
}

// DrainBody reads all of b to memory and then returns two equivalent
// ReadClosers yielding the same bytes.
//
// It returns an error if the initial slurp of all bytes fails. It does not attempt
// to make the returned ReadClosers have identical error-matching behavior.
func DrainBody(b io.ReadCloser) (rb []byte, nopb io.ReadCloser, err error) {
	if b == http.NoBody {
		// No copying needed. Preserve the magic sentinel meaning of NoBody.
		return nil, http.NoBody, nil
	}
	var buf bytes.Buffer
	if _, err = buf.ReadFrom(b); err != nil {
		return nil, b, err
	}
	if err = b.Close(); err != nil {
		return nil, b, err
	}
	return buf.Bytes(), ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}
