package fetch

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/beanscc/fetch/binding"
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

// BindWith bind http.Response
func (r *response) BindWith(obj interface{}, b binding.Binding) error {
	if r.err != nil {
		return r.err
	}

	return b.Bind(r.resp, obj)
}

func (r *response) BindBody(obj interface{}, b binding.BindingBody) error {
	if r.err != nil {
		return r.err
	}

	return b.BindBody(r.body, obj)
}

// BindJSON bind resp body with json
func (r *response) BindJSON(obj interface{}) error {
	return r.BindBody(obj, &binding.Json{})
}

func (r *response) Resp() (*http.Response, error) {
	return r.resp, r.err
}

func (r *response) Body() ([]byte, error) {
	return r.body, r.err
}

func (r *response) Error() error {
	return r.err
}

// NopCloserRespBody 返回一个不需要 close 的 io.ReadCloser
func NopCloserRespBody(b io.ReadCloser) (io.ReadCloser, error) {
	if b == http.NoBody {
		return http.NoBody, nil
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(b); err != nil {
		return b, err
	}
	if err := b.Close(); err != nil {
		return b, err
	}

	return ioutil.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

// drainBody reads all of b to memory and then returns two equivalent
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