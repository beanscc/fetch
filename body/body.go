package body

import (
	"io"
)

// Body 构造请求的body
type Body interface {
	// Body 构造http请求body
	Body() (io.Reader, error)
	// ContentType 返回 body 体结构相应的 Header content-type 类型
	ContentType() string
}

// Body 接口实现检查
var (
	_ Body = &JSON{}
	_ Body = &XML{}
	_ Body = &Form{}
	_ Body = &MultipartForm{}
)
