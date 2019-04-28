package fetch

import (
	"net/http"

	"github.com/beanscc/fetch/binding"
)

type response struct {
	resp *http.Response
	body []byte
	err  error
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

func (r *response) Resp() (*http.Response, error) {
	return r.resp, r.err
}

func (r *response) Body() ([]byte, error) {
	return r.body, r.err
}

func (r *response) Error() error {
	return r.err
}
