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
	// v 应该是一个指针对象
	Bind(resp *http.Response, body []byte, v interface{}) error
}

var (
	// Binding 接口实现检查
	_ Binding = &JSON{}
	_ Binding = &XML{}
)
