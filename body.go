package fetch

import (
	"io"
)

// Body 构造请求的body
type Body interface {
	// Body 构造http请求body
	Body() (io.Reader, error)
	// ContentType 返回 body 体相应的 content-type
	ContentType() string
}

// Body 接口实现检查
var (
	_ Body = &JsonBody{}
	_ Body = &JsonStrBody{}
	_ Body = &FormDataBody{}
	_ Body = &XWWWFormURLEncodedBody{}
)
