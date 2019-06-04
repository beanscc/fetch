package binding

import (
	"net/http"
)

// Binding 解析 http.Response 的接口定义
type Binding interface {
	// Name Binding对象的名称
	Name() string

	// Bind 解析 http.Response 的func
	// resp 可能是 nil，也可能由于中间件在读取完resp.Body 后未还原body，导致 body为空
	// out 应该是一个指针对象
	Bind(resp *http.Response, out interface{}) error
}

// BindingBody adds BindBody method to Binding. BindBody is similar with Bind,
// but it reads the body from supplied bytes instead of resp.Body.
type BindingBody interface {
	Binding
	// 解析 http 响应 body 消息的func
	BindBody([]byte, interface{}) error
}
