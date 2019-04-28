package binding

import (
	"net/http"
)

type Binding interface {
	Name() string
	Bind(resp *http.Response, out interface{}) error
}

// BindingBody adds BindBody method to Binding. BindBody is similar with Bind,
// but it reads the body from supplied bytes instead of resp.Body.
type BindingBody interface {
	Binding
	BindBody([]byte, interface{}) error
}
